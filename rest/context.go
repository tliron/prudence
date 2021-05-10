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
	RequestContext *fasthttp.RequestCtx
	Path           string
	Method         string
	Variables      map[string]string
	ContentType    string
	LastModified   time.Time
	ETag           string
	MaxAge         int
	Writer         io.Writer
	Log            logging.Logger
}

func NewContext(requestContext *fasthttp.RequestCtx) *Context {
	return &Context{
		RequestContext: requestContext,
		Path:           util.BytesToString(requestContext.Path()[1:]), // without initial "/"
		Method:         util.BytesToString(requestContext.Method()),
		Variables:      make(map[string]string),
		MaxAge:         -1,
		Writer:         requestContext,
		Log:            log,
	}
}

func (self *Context) AutoETag() {
	self.Log.Info("auto ETag")
	self.Writer = NewETagBuffer(self.Writer)
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
		RequestContext: self.RequestContext,
		Path:           self.Path,
		Method:         self.Method,
		Variables:      variables,
		ContentType:    self.ContentType,
		ETag:           self.ETag,
		MaxAge:         self.MaxAge,
		Writer:         self.Writer,
		Log:            self.Log,
	}
}
