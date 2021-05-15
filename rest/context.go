package rest

import (
	"bytes"
	"io"
	"time"

	"github.com/tliron/kutil/ard"
	"github.com/tliron/kutil/js"
	"github.com/tliron/kutil/logging"
	"github.com/tliron/kutil/util"
	"github.com/valyala/fasthttp"
)

//
// Context
//

type Context struct {
	Log         logging.Logger
	Variables   ard.StringMap
	Path        string
	Method      string
	Query       map[string]string
	ContentType string
	CharSet     string
	Language    string

	CacheDuration float64 // seconds
	CacheKey      string
	Signature     string
	WeakSignature bool
	LastModified  time.Time

	context *fasthttp.RequestCtx
	writer  io.Writer
}

func NewContext(context *fasthttp.RequestCtx) *Context {
	return &Context{
		Log:       log,
		Variables: make(ard.StringMap),
		Path:      util.BytesToString(context.Path()[1:]), // without initial "/"
		Method:    util.BytesToString(context.Method()),
		Query:     GetQuery(context),
		context:   context,
		writer:    context,
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

func (self *Context) EndSignature() {
	if hashWriter, ok := self.writer.(*HashWriter); ok {
		self.Log.Debug("end signature")
		self.Signature = hashWriter.Hash()
		self.writer = hashWriter.Writer
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

// io.Writer
func (self *Context) Write(b []byte) (int, error) {
	return self.writer.Write(b)
}

func (self *Context) Copy() *Context {
	variables := ard.Copy(self.Variables).(ard.StringMap)

	return &Context{
		Log:           self.Log,
		Variables:     variables,
		Path:          self.Path,
		Method:        self.Method,
		Query:         self.Query,
		ContentType:   self.ContentType,
		CharSet:       self.CharSet,
		Language:      self.Language,
		CacheDuration: self.CacheDuration,
		CacheKey:      self.CacheKey,
		Signature:     self.Signature,
		WeakSignature: self.WeakSignature,
		LastModified:  self.LastModified,
		context:       self.context,
		writer:        self.writer,
	}
}

func (self *Context) Embed(hook *js.Hook) {
	// Try cache
	if self.CacheKey != "" {
		if cacheEntry, ok := CacheLoad(self); ok {
			if self.context.IsHead() {
				// HEAD doesn't care if the cacheEntry doesn't have a body
				cacheEntry.Write(self)
				return
			} else {
				if len(cacheEntry.Body) == 0 {
					self.Log.Debugf("ignoring cache with no body: %s", self.Path)
				} else {
					cacheEntry.Write(self)
					return
				}
			}
		}
	}

	buffer := bytes.NewBuffer(nil)
	writer := self.writer
	self.writer = buffer

	hook.Call(nil, self)

	body := buffer.Bytes()

	// To cache
	if (self.CacheDuration > 0.0) && (self.CacheKey != "") {
		CacheStoreBody(self, "", body)
	}

	self.writer = writer
	self.Write(body)
}
