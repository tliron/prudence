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
	Log     logging.Logger
	Scratch map[string]interface{}

	Path          string
	Method        string
	Query         map[string]string
	Variables     map[string]string
	ContentType   string
	CharSet       string
	Language      string
	CacheDuration float64 // seconds
	CacheKey      string
	ETag          string
	WeakETag      bool
	LastModified  time.Time

	context *fasthttp.RequestCtx
	writer  io.Writer
}

func NewContext(context *fasthttp.RequestCtx) *Context {
	query := make(map[string]string)
	context.QueryArgs().VisitAll(func(key []byte, value []byte) {
		log.Debugf("query: %s = %s", key, value)
		query[util.BytesToString(key)] = util.BytesToString(value)
	})

	return &Context{
		Log:       log,
		Scratch:   make(map[string]interface{}),
		Path:      util.BytesToString(context.Path()[1:]), // without initial "/"
		Method:    util.BytesToString(context.Method()),
		Query:     query,
		Variables: make(map[string]string),
		context:   context,
		writer:    context,
	}
}

// Calculating an ETag from the body is not that great. It saves bandwidth but not computing
// resources, as we still need to generate the body in order to calculate the ETag. Ideally,
// the ETag should be based on the data sources used to generate the page.
//
// https://www.mnot.net/blog/2007/08/07/etags
// http://www.tbray.org/ongoing/When/200x/2007/07/31/Design-for-the-Web
func (self *Context) StartETag() {
	if _, ok := self.writer.(*HashWriter); !ok {
		self.Log.Info("start ETag")
		self.writer = NewHashWriter(self.writer)
	}
	// TODO: if a low priority part of the page changes then set WeakETag=true
}

func (self *Context) EndETag() {
	if hashWriter, ok := self.writer.(*HashWriter); ok {
		self.Log.Info("end ETag")
		self.ETag = hashWriter.Hash()
		self.writer = hashWriter.Writer
	}
}

func (self *Context) RenderETag() string {
	if self.ETag != "" {
		if self.WeakETag {
			return "W/\"" + self.ETag + "\""
		} else {
			return "\"" + self.ETag + "\""
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
	variables := make(map[string]string)
	for key, value := range self.Variables {
		variables[key] = value
	}

	return &Context{
		Log:           self.Log,
		Scratch:       make(map[string]interface{}),
		Path:          self.Path,
		Method:        self.Method,
		Query:         self.Query,
		Variables:     variables,
		ContentType:   self.ContentType,
		CharSet:       self.CharSet,
		Language:      self.Language,
		CacheDuration: self.CacheDuration,
		CacheKey:      self.CacheKey,
		ETag:          self.ETag,
		WeakETag:      self.WeakETag,
		LastModified:  self.LastModified,
		context:       self.context,
		writer:        self.writer,
	}
}
