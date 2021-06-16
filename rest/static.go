package rest

import (
	"net/http"
	"path/filepath"

	"github.com/tliron/kutil/ard"
	"github.com/tliron/kutil/js"
	urlpkg "github.com/tliron/kutil/url"
	"github.com/tliron/prudence/platform"
)

func init() {
	platform.RegisterType("Static", CreateStatic)
}

//
// Static
//

type Static struct {
	Root string
}

func NewStatic(root string, indexes []string) *Static {
	return &Static{
		Root: root,
	}
}

// CreateFunc signature
func CreateStatic(config ard.StringMap, context *js.Context) (interface{}, error) {
	config_ := ard.NewNode(config)

	root, _ := config_.Get("root").String(false)
	if rootUrl, err := context.Resolve(root, true); err == nil {
		root = rootUrl.(*urlpkg.FileURL).Path
	} else {
		return nil, err
	}
	indexes := platform.AsStringList(config_.Get("indexes").Data)

	return NewStatic(root, indexes), nil
}

// Handler interface
// HandleFunc signature
func (self *Static) Handle(context *Context) bool {
	path := filepath.Join(self.Root, context.Path)
	http.ServeFile(NewResponseWriterWrapper(context), context.Request.Direct, path)
	if context.Response.Status != http.StatusNotFound {
		context.Response.Bypass = true
		return true
	} else {
		return false
	}
}

// https://stackoverflow.com/a/47286697

type ResponseWriterWrapper struct {
	http.ResponseWriter
	context *Context
}

func NewResponseWriterWrapper(context *Context) *ResponseWriterWrapper {
	return &ResponseWriterWrapper{
		ResponseWriter: context.Response.Direct,
		context:        context,
	}
}

func (self *ResponseWriterWrapper) WriteHeader(status int) {
	self.context.Response.Status = status
	self.ResponseWriter.WriteHeader(status)
}

func (self *ResponseWriterWrapper) Write(p []byte) (int, error) {
	if self.context.Response.Status != http.StatusNotFound {
		return self.ResponseWriter.Write(p)
	} else {
		return len(p), nil
	}
}
