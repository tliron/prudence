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
	Writer io.Writer
	Log    logging.Logger

	Context       *fasthttp.RequestCtx
	Path          string
	Method        string
	Variables     map[string]string
	ContentType   string
	LastModified  time.Time
	ETag          string
	WeakETag      bool
	CacheDuration float64 // seconds
}

func NewContext(context *fasthttp.RequestCtx) *Context {
	return &Context{
		Writer:    context,
		Log:       log,
		Context:   context,
		Path:      util.BytesToString(context.Path()[1:]), // without initial "/"
		Method:    util.BytesToString(context.Method()),
		Variables: make(map[string]string),
	}
}

func (self *Context) AutoETag() {
	self.Log.Info("auto ETag")
	self.Writer = NewHashWriter(self.Writer)
	// TODO: if a low priority part of the page changes then set WeakETag=true
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
		Context:       self.Context,
		Path:          self.Path,
		Method:        self.Method,
		Variables:     variables,
		ContentType:   self.ContentType,
		LastModified:  self.LastModified,
		ETag:          self.ETag,
		WeakETag:      self.WeakETag,
		CacheDuration: self.CacheDuration,
	}
}
