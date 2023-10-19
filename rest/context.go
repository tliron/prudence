package rest

import (
	"errors"
	"fmt"
	"net/http"
	urlpkg "net/url"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/tliron/commonjs-goja/api"
	"github.com/tliron/commonlog"
	"github.com/tliron/go-scriptlet/jst"
)

//
// Context
//

type Context struct {
	*jst.Context

	Id       uint64
	Request  *Request
	Response *Response

	Name  string
	Log   commonlog.Logger
	Debug bool

	Done    bool
	Created bool
	Async   bool

	CacheDuration float64 // seconds
	CacheKey      string
	CacheGroups   []string
}

var requestId atomic.Uint64

func NewContext(responseWriter http.ResponseWriter, request *http.Request, log commonlog.Logger) *Context {
	id := requestId.Add(1)
	request_ := NewRequest(request)
	response := NewResponse(responseWriter)
	return &Context{
		Context:  jst.NewContext(response.Buffer, nil),
		Id:       id,
		Request:  request_,
		Response: response,
		Log: commonlog.NewKeyValueLogger(log,
			"_scope", "request",
			"request", id,
			"host", request_.Host,
			"port", request_.Port,
			"path", request_.Path,
		),
	}
}

func (self *Context) Clone() *Context {
	return &Context{
		Context:  self.Context.Clone(),
		Id:       self.Id,
		Request:  self.Request, //.Clone(),
		Response: self.Response,
		Name:     self.Name,
		Log:      self.Log,
		Debug:    self.Debug,
	}
}

func (self *Context) AppendName(name string, alwaysClone bool) *Context {
	if name == "" {
		if alwaysClone {
			return self.Clone()
		} else {
			return self
		}
	} else {
		restContext := self.Clone()
		if restContext.Name == "" {
			restContext.Name = name
		} else {
			restContext.Name += "." + name
		}
		restContext.Log = commonlog.NewKeyValueLogger(self.Log, "resource", restContext.Name)
		return restContext
	}
}

// Ends request handling (via a panic) and flushes the current response.
func (self *Context) End() {
	panic(EndRequest)
}

// Ends request handling (via a panic) and sends a redirect header to the
// client. If status is 0 will default to 302 (Found).
//
// See: https://developer.mozilla.org/en-US/docs/Web/HTTP/Redirections
func (self *Context) Redirect(url string, status int) error {
	if status == 0 {
		status = http.StatusFound // 302
	} else if (status < 300) || (status >= 400) {
		return fmt.Errorf("not a redirect code: %d", status)
	}

	self.Response.Reset()
	self.Response.Status = status
	self.Response.Header.Set(HeaderLocation, url)
	self.Log.Infof("redirect %d: %s", status, url)
	panic(EndRequest)
}

// If the request URL path does not end in a slash will call
// [Context.Redirect] with an appended slash. The URL query
// will be preserved.
//
// Does nothing if the URL path already ends in a slash.
func (self *Context) RedirectTrailingSlash(status int) error {
	if strings.HasSuffix(self.Request.Direct.URL.Path, "/") {
		// Nothing to do
		return nil
	}

	url := urlpkg.URL{
		Path:     self.Request.Direct.URL.Path + "/",
		RawQuery: self.Request.Direct.URL.RawQuery,
	}

	return self.Redirect(url.String(), status)
}

// Ends request handling (via a panic) and sends a 500 status to the
// client. If err is provided will log the error.
//
// Note that no content is set to the client with this call. If you
// wish to send content, do so before calling this function.
func (self *Context) InternalServerError(err error) {
	if err == nil {
		self.Log.Error("InternalServerError")
	} else {
		self.Log.Errorf("InternalServerError: %s", err.Error())
	}
	self.Response.Status = http.StatusInternalServerError
	panic(EndRequest)
}

// Encodes and writes the value. If format is empty, will try to set the format
// according to the value of [Response.ContentType]. Supported formats are "yaml",
// "json", "xjson", "xml", "cbor", "messagepack", and "go". The "cbor" and
// "messsagepack" formats will be encoded in base64 and will ignore the indent argument.
func (self *Context) Transcribe(value any, format string, indent string) error {
	if format == "" {
		switch self.Response.ContentType {
		case "application/yaml":
			format = "yaml"
		case "application/json":
			format = "json"
		case "application/xml":
			format = "xml"
		case "application/cbor":
			format = "cbor"
		case "application/msgpack":
			format = "messagepack"
		default:
			return fmt.Errorf("cannot determine format from content type: %s", self.Response.ContentType)
		}
	}

	var transcriber api.Transcribe
	return transcriber.Write(self.Writer, value, format, indent)
}

// After calling this function, subsequent calls to [Context.Write] will be accumulated
// into a signature calculation. Call [Context.EndSignature] when done.
//
// Calculating a signature from the body is not that great. It saves bandwidth but not computing
// resources, as we still need to generate the body in order to calculate the signature. Ideally,
// the signature should be based on the data sources used to generate the page.
//
// https://www.mnot.net/blog/2007/08/07/etags
// http://www.tbray.org/ongoing/When/200x/2007/07/31/Design-for-the-Web
func (self *Context) StartSignature() {
	if _, ok := self.Writer.(*HashWriter); !ok {
		self.Writer = NewHashWriter(self.Writer)
	}
}

// Must be called subsequently to [Context.StartSignature]. The calculated signature
// will be put into the context's [Response.Signature].
func (self *Context) EndSignature() error {
	if hashWriter, ok := self.Writer.(*HashWriter); ok {
		self.Response.Signature = hashWriter.Hash()
		self.Writer = hashWriter.writer
		return nil
	} else {
		return errors.New("did not call startSignature()")
	}
}

// Utils

func (self *Context) caching() bool {
	return (self.CacheDuration > 0.0) && (self.CacheKey != "")
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
		if clientTimestamp, ok := GetTimeHeader(HeaderIfModifiedSince, self.Request.Header); ok {
			// modified = server > client
			// not modified = client <= server
			if !serverTimestamp.Truncate(time.Second).After(clientTimestamp) {
				self.Response.Status = http.StatusNotModified
				self.Log.Debug("not modified: Last-Modified")
				return true
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
		maxAge := int64(self.CacheDuration)
		self.Response.Header.Set(HeaderCacheControl, "max-age="+strconv.FormatInt(maxAge, 10))
	}
}
