package rest

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/dop251/goja"
	"github.com/tliron/kutil/ard"
	"github.com/tliron/kutil/logging"
	"github.com/tliron/kutil/util"
	"github.com/tliron/prudence/platform"
	"github.com/valyala/fasthttp"
)

//
// Context
//

type Context struct {
	Context     *fasthttp.RequestCtx
	Log         logging.Logger
	Debug       bool
	Variables   ard.StringMap
	Path        string
	Method      string
	Query       map[string][]string
	ContentType string
	CharSet     string
	Language    string
	Done        bool
	Created     bool
	Async       bool

	CacheDuration float64 // seconds
	CacheKey      string
	CacheGroups   []string
	Signature     string
	WeakSignature bool
	Timestamp     time.Time

	writer io.Writer
}

func NewContext(context *fasthttp.RequestCtx) *Context {
	return &Context{
		Context:   context,
		Log:       log,
		Variables: make(ard.StringMap),
		Path:      util.BytesToString(context.Path()[1:]), // without initial "/"
		Method:    util.BytesToString(context.Method()),
		Query:     GetQuery(context),
		writer:    context,
	}
}

func (self *Context) Copy() *Context {
	variables := ard.Copy(self.Variables).(ard.StringMap)
	cacheGroups := make([]string, len(self.CacheGroups))
	copy(cacheGroups, self.CacheGroups)

	return &Context{
		Context:       self.Context,
		Log:           self.Log,
		Debug:         self.Debug,
		Variables:     variables,
		Path:          self.Path,
		Method:        self.Method,
		Query:         self.Query,
		ContentType:   self.ContentType,
		CharSet:       self.CharSet,
		Language:      self.Language,
		Done:          self.Done,
		Created:       self.Created,
		Async:         self.Async,
		CacheDuration: self.CacheDuration,
		CacheKey:      self.CacheKey,
		CacheGroups:   cacheGroups,
		Signature:     self.Signature,
		WeakSignature: self.WeakSignature,
		Timestamp:     self.Timestamp,
		writer:        self.writer,
	}
}

func (self *Context) StartCapture(name string) {
	self.writer = NewCaptureWriter(self.writer, name, func(name string, value string) {
		self.Variables[name] = value
	})
}

func (self *Context) EndCapture() error {
	if captureWriter, ok := self.writer.(*CaptureWriter); ok {
		self.Log.Debug("end capture")
		err := captureWriter.Close()
		self.writer = captureWriter.GetWrappedWriter()
		return err
	} else {
		return errors.New("did not call startCapture()")
	}
}

func (self *Context) StartRender(renderer string, hasGetRelativeURL platform.HasGetRelativeURL) error {
	if renderWriter, err := NewRenderWriter(self.writer, renderer, hasGetRelativeURL.GetRelativeURL); err == nil {
		self.Log.Debugf("start render: %s", renderer)
		self.writer = renderWriter
		return nil
	} else {
		return err
	}
}

func (self *Context) EndRender() error {
	if renderWriter, ok := self.writer.(*RenderWriter); ok {
		self.Log.Debug("end render")
		err := renderWriter.Close()
		self.writer = renderWriter.GetWrappedWriter()
		return err
	} else {
		return errors.New("did not call startRender()")
	}
}

// Calculating a signature from the body is not that great. It saves bandwidth but not computing
// resources, as we still need to generate the body in order to calculate the signature. Ideally,
// the signature should be based on the data sources used to generate the page.
//
// https://www.mnot.net/blog/2007/08/07/etags
// http://www.tbray.org/ongoing/When/200x/2007/07/31/Design-for-the-Web
func (self *Context) StartSignature() {
	if _, ok := self.writer.(*HashWriter); !ok {
		self.Log.Debug("start signature")
		self.writer = NewHashWriter(self.writer)
	}
}

func (self *Context) EndSignature() error {
	if hashWriter, ok := self.writer.(*HashWriter); ok {
		self.Log.Debug("end signature")
		self.Signature = hashWriter.Hash()
		self.writer = hashWriter.writer
		return nil
	} else {
		return errors.New("did not call startSignature()")
	}
}

func (self *Context) ETag() string {
	if self.Signature != "" {
		if self.WeakSignature {
			return "W/\"" + self.Signature + "\""
		} else {
			return "\"" + self.Signature + "\""
		}
	} else {
		return ""
	}
}

func (self *Context) Error(err error) {
	self.Log.Errorf("%s", err)
	self.Context.SetStatusCode(fasthttp.StatusInternalServerError)
}

func (self *Context) Request() string {
	return util.BytesToString(self.Context.Request.Body())
}

// io.Writer
func (self *Context) Write(b []byte) (int, error) {
	return self.writer.Write(b)
}

func (self *Context) WriteString(s string) (int, error) {
	return self.writer.Write(util.StringToBytes(s))
}

func (self *Context) Embed(present interface{}, runtime *goja.Runtime) error {
	// Try cache
	if self.CacheKey != "" {
		if cacheKey, cacheEntry, ok := CacheLoad(self); ok {
			if len(cacheEntry.Body) == 0 {
				self.Log.Debugf("ignoring cache with no body: %s", self.Path)
			} else {
				changed, _, err := CacheEntryWrite(cacheEntry, self)
				if err != nil {
					self.Log.Errorf("%s", err.Error())
				}
				if changed {
					CacheUpdate(cacheKey, cacheEntry)
				}
				return nil
			}
		}
	}

	buffer := bytes.NewBuffer(nil)
	writer := self.writer
	self.writer = buffer

	if function, ok := present.(JavaScriptFunc); ok {
		CallJavaScript(runtime, function, self)
	} else {
		return fmt.Errorf("not a function: %T", present)
	}

	self.EndSignature()

	body := buffer.Bytes()

	// To cache
	if (self.CacheDuration > 0.0) && (self.CacheKey != "") {
		CacheStoreBody(self, platform.EncodingTypePlain, body)
	}

	self.writer = writer
	self.Write(body)

	return nil
}

func (self *Context) unwrapWriters() {
	for true {
		if wrappedWriter, ok := self.writer.(WrappingWriter); ok {
			if err := wrappedWriter.Close(); err != nil {
				self.Error(err)
			}
			self.writer = wrappedWriter.GetWrappedWriter()
		} else {
			break
		}
	}
}

func (self *Context) setContentType() {
	// Content-Type
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Type
	if self.ContentType != "" {
		if self.CharSet != "" {
			self.Context.SetContentType(self.ContentType + ";charset=" + self.CharSet)
		} else {
			self.Context.SetContentType(self.ContentType)
		}
	}
}

func (self *Context) addETag() {
	// ETag
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/ETag
	if eTag := self.ETag(); eTag != "" {
		AddETag(self.Context, eTag)
	}
}

func (self *Context) setLastModified() {
	// Last-Modified
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Last-Modified
	if !self.Timestamp.IsZero() {
		self.Context.Response.Header.SetLastModified(self.Timestamp)
	}
}

func (self *Context) setCacheControl() {
	// Cache-Control
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Cache-Control
	if self.CacheDuration < 0.0 {
		// Negative means don't store and *also* invalidate the existing client cache
		AddCacheControl(self.Context, "no-store,max-age=0")
	} else if self.CacheDuration > 0.0 {
		// Match client-side caching with server-side caching
		maxAge := int(self.CacheDuration)
		AddCacheControl(self.Context, fmt.Sprintf("max-age=%d", maxAge))
	}
}

func (self *Context) isNotModified() bool {
	// If-None-Match
	// (Has precedence over If-Modified-Since)
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/If-None-Match
	if IfNoneMatch(self.Context, self.ETag()) {
		self.Context.NotModified()
		self.Log.Debug("not modified: ETag")
		return true
	}

	// If-Modified-Since
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/If-Modified-Since
	if !self.Timestamp.IsZero() {
		if !self.Context.IfModifiedSince(self.Timestamp) {
			self.Context.NotModified()
			self.Log.Debug("not modified: Last-Modified")
			return true
		}
	}

	return false
}
