package rest

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/dop251/goja"
	"github.com/tliron/kutil/ard"
	"github.com/tliron/kutil/format"
	"github.com/tliron/kutil/js"
	"github.com/tliron/kutil/logging"
	"github.com/tliron/kutil/util"
	"github.com/tliron/prudence/platform"
)

//
// Context
//

type Context struct {
	Request  *Request
	Response *Response

	Log   logging.Logger
	Name  string
	Debug bool

	Path      string
	Variables ard.StringMap

	Done    bool
	Created bool
	Async   bool

	CacheDuration float64 // seconds
	CacheKey      string
	CacheGroups   []string

	writer io.Writer
}

func NewContext(responseWriter http.ResponseWriter, request *http.Request) *Context {
	self := Context{
		Request:   NewRequest(request),
		Response:  NewResponse(responseWriter),
		Log:       log,
		Path:      request.URL.Path[1:], // without initial "/"
		Variables: make(ard.StringMap),
	}
	self.writer = self.Response.Buffer
	return &self
}

func (self *Context) AppendName(name string) *Context {
	if name == "" {
		return self
	} else {
		context := self.Copy()
		if context.Name == "" {
			context.Name = name
		} else {
			context.Name += "." + name
		}
		context.Log = logging.NewSubLogger(log, context.Name)
		return context
	}
}

func (self *Context) Copy() *Context {
	return &Context{
		Request:   self.Request,
		Response:  self.Response,
		Log:       self.Log,
		Name:      self.Name,
		Debug:     self.Debug,
		Path:      self.Path,
		Variables: ard.Copy(self.Variables).(ard.StringMap),
		writer:    self.writer,
	}
}

func (self *Context) Redirect(url string, status int) error {
	// See: https://developer.mozilla.org/en-US/docs/Web/HTTP/Redirections

	if status == 0 {
		status = http.StatusFound // 302
	} else if (status < 300) || (status >= 400) {
		return fmt.Errorf("not a redirect code: %d", status)
	}

	self.Response.Reset()
	self.Response.Status = status
	self.Response.Header.Set(HeaderLocation, url)
	return nil
}

func (self *Context) InternalServerError(err error) {
	self.Log.Errorf("%s", err)
	self.Response.Status = http.StatusInternalServerError
}

// io.Writer
func (self *Context) Write(b []byte) (int, error) {
	return self.writer.Write(b)
}

func (self *Context) WriteString(s string) (int, error) {
	return self.writer.Write(util.StringToBytes(s))
}

func (self *Context) WriteJson(value ard.Value, indent string) (int, error) {
	if s, err := format.Encode(value, "json", indent, false); err == nil {
		return self.WriteString(s)
	} else {
		return 0, err
	}
}

func (self *Context) WriteYaml(value ard.Value, indent string) (int, error) {
	if s, err := format.Encode(value, "yaml", indent, false); err == nil {
		return self.WriteString(s)
	} else {
		return 0, err
	}
}

func (self *Context) Embed(function goja.FunctionCall, runtime *goja.Runtime) goja.Value {
	var present js.JavaScriptFunc
	if len(function.Arguments) > 0 {
		var ok bool
		present_ := function.Arguments[0].Export()
		if bind, ok := present_.(js.Bind); ok {
			var err error
			var jsContext *js.Context
			if present_, jsContext, err = bind.Unbind(); err == nil {
				runtime = jsContext.Environment.Runtime
			} else {
				panic(runtime.NewGoError(err))
			}
		}

		if present, ok = present_.(js.JavaScriptFunc); !ok {
			panic(runtime.NewGoError(fmt.Errorf("\"present\" not a function: %T", present_)))
		}
	} else {
		panic(runtime.NewGoError(errors.New("missing \"present\" argument")))
	}

	// Try cache
	if self.CacheKey != "" {
		if key, cached, ok := self.LoadCachedRepresentation(); ok {
			if len(cached.Body) == 0 {
				self.Log.Debugf("ignoring cache with no body: %s", self.Path)
			} else {
				changed, _, err := self.WriteCachedRepresentation(cached)
				if err != nil {
					self.Log.Errorf("%s", err.Error())
				}
				if changed {
					cached.Update(key)
				}
				return nil
			}
		}
	}

	buffer := bytes.NewBuffer(nil)
	writer := self.writer
	self.writer = buffer

	js.Call(runtime, present, self)

	self.flushWriters()

	body := buffer.Bytes()

	// To cache
	if (self.CacheDuration > 0.0) && (self.CacheKey != "") {
		self.StoreCachedRepresentationFromBody(platform.EncodingTypeIdentity, body)
	}

	self.writer = writer
	self.Write(body)

	return nil
}

func (self *Context) flushWriters() {
	for true {
		if wrappedWriter, ok := self.writer.(WrappingWriter); ok {
			if err := wrappedWriter.Close(); err != nil {
				self.InternalServerError(err)
			}
			self.writer = wrappedWriter.GetWrappedWriter()
		} else {
			break
		}
	}
}

func (self *Context) isNotModified(fromHeader bool) bool {
	// If-None-Match
	// (Has precedence over If-Modified-Since)
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/If-None-Match
	if serverETag := self.Response.eTag(fromHeader); serverETag != "" {
		if clientETag := self.Request.Header.Get(HeaderIfNoneMatch); clientETag != "" {
			if clientETag == serverETag {
				self.Response.Status = http.StatusNotModified
				self.Log.Debug("not modified: ETag")
				return true
			}
		}
	}

	// If-Modified-Since
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/If-Modified-Since
	if serverTimestamp := self.Response.lastModified(fromHeader); !serverTimestamp.IsZero() {
		if ifModifiedSince := self.Request.Header.Get(HeaderIfModifiedSince); ifModifiedSince != "" {
			if clientTimestamp, err := http.ParseTime(ifModifiedSince); err == nil {
				// modified = server > client
				// not modified = client <= server
				if !serverTimestamp.After(clientTimestamp) {
					self.Response.Status = http.StatusNotModified
					self.Log.Debug("not modified: Last-Modified")
					return true
				}
			}
		}
	}

	return false
}

func (self *Context) setCacheControl() {
	// Cache-Control
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Cache-Control
	if self.CacheDuration < 0.0 {
		// Negative means don't store and *also* invalidate the existing client cache
		self.Response.Header.Set(HeaderCacheControl, "no-store,max-age=0")
	} else if self.CacheDuration > 0.0 {
		// Match client-side caching with server-side caching
		maxAge := int(self.CacheDuration)
		self.Response.Header.Set(HeaderCacheControl, fmt.Sprintf("max-age=%d", maxAge))
	}
}
