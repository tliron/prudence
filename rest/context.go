package rest

import (
	"bytes"
	"errors"
	"io"
	"time"

	"github.com/tliron/kutil/ard"
	"github.com/tliron/kutil/js"
	"github.com/tliron/kutil/logging"
	"github.com/tliron/kutil/util"
	"github.com/tliron/prudence/platform"
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
	Query       map[string][]string
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
		self.writer = renderWriter.Writer
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
	self.context.SetStatusCode(fasthttp.StatusInternalServerError)
}

// io.Writer
func (self *Context) Write(b []byte) (int, error) {
	return self.writer.Write(b)
}

func (self *Context) WriteString(s string) (int, error) {
	return self.writer.Write(util.StringToBytes(s))
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
		if cacheKey, cacheEntry, ok := CacheLoad(self); ok {
			if len(cacheEntry.Body) == 0 {
				self.Log.Debugf("ignoring cache with no body: %s", self.Path)
			} else {
				changed, _, err := CacheEntryWritePlain(cacheEntry, self)
				if err != nil {
					self.Log.Errorf("%s", err.Error())
				}
				if changed {
					CacheUpdate(cacheKey, cacheEntry)
				}
				return
			}
		}
	}

	buffer := bytes.NewBuffer(nil)
	writer := self.writer
	self.writer = buffer

	hook.Call(nil, self)

	self.EndSignature()

	body := buffer.Bytes()

	// To cache
	if (self.CacheDuration > 0.0) && (self.CacheKey != "") {
		CacheStoreBody(self, platform.EncodingTypePlain, body)
	}

	self.writer = writer
	self.Write(body)
}
