package rest

import (
	"io"
	"time"

	"github.com/tliron/kutil/logging"
	"github.com/tliron/kutil/util"
	"github.com/valyala/fasthttp"
)

//
// Context
//

type Context struct {
	Writer  io.Writer
	Log     logging.Logger
	Scratch map[string]interface{}

	Path          string
	Method        string
	Variables     map[string]string
	ContentType   string
	CacheDuration float64 // seconds
	CacheKey      string
	ETag          string
	WeakETag      bool
	LastModified  time.Time

	context *fasthttp.RequestCtx
}

func NewContext(context *fasthttp.RequestCtx) *Context {
	return &Context{
		Writer:    context,
		Log:       log,
		Scratch:   make(map[string]interface{}),
		Path:      util.BytesToString(context.Path()[1:]), // without initial "/"
		Method:    util.BytesToString(context.Method()),
		Variables: make(map[string]string),
		context:   context,
	}
}

// Calculating an ETag from the body is not that great. It saves bandwidth but not computing
// resources, as we still need to generate the body in order to calculate the ETag. Ideally,
// the ETag should be based on the data sources used to generate the page.
//
// https://www.mnot.net/blog/2007/08/07/etags
// http://www.tbray.org/ongoing/When/200x/2007/07/31/Design-for-the-Web
func (self *Context) StartETag() {
	if _, ok := self.Writer.(*HashWriter); !ok {
		self.Log.Info("start ETag")
		self.Writer = NewHashWriter(self.Writer)
	}
	// TODO: if a low priority part of the page changes then set WeakETag=true
}

func (self *Context) EndETag() {
	if hashWriter, ok := self.Writer.(*HashWriter); ok {
		self.Log.Info("end ETag")
		self.ETag = hashWriter.Hash()
		self.Writer = hashWriter.Writer
	}
}

// io.Writer
func (self *Context) Write(b []byte) (int, error) {
	return self.Writer.Write(b)
}

func (self *Context) Copy() *Context {
	variables := make(map[string]string)
	for key, value := range self.Variables {
		variables[key] = value
	}

	return &Context{
		Writer:        self.Writer,
		Log:           self.Log,
		Scratch:       make(map[string]interface{}),
		Path:          self.Path,
		Method:        self.Method,
		Variables:     variables,
		ContentType:   self.ContentType,
		CacheDuration: self.CacheDuration,
		CacheKey:      self.CacheKey,
		ETag:          self.ETag,
		WeakETag:      self.WeakETag,
		LastModified:  self.LastModified,
		context:       self.context,
	}
}
