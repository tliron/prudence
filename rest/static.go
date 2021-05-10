package rest

import (
	"github.com/tliron/kutil/ard"
	urlpkg "github.com/tliron/kutil/url"
	"github.com/tliron/prudence/js/common"
	"github.com/valyala/fasthttp"
)

func init() {
	Register("static", CreateStatic)
}

//
// Static
//

type Static struct {
	RequestHandler fasthttp.RequestHandler
}

func NewStatic(root string, indexes []string) *Static {
	fs := fasthttp.FS{
		Root:               root,
		IndexNames:         indexes,
		GenerateIndexPages: true,
		Compress:           true, // TODO ???
		CompressBrotli:     true,
		PathRewrite: func(context *fasthttp.RequestCtx) []byte {
			log.Infof(">>> %s", context.Request.Header.Peek("__path"))
			return context.Request.Header.Peek("__path")
		},
	}

	return &Static{
		RequestHandler: fs.NewRequestHandler(),
	}
}

// CreateFunc signature
func CreateStatic(config ard.StringMap, getRelativeURL common.GetRelativeURL) (interface{}, error) {
	config_ := ard.NewNode(config)
	root, _ := config_.Get("root").String(false)
	if rootUrl, err := getRelativeURL(root); err == nil {
		root = rootUrl.(*urlpkg.FileURL).Path
	} else {
		return nil, err
	}
	indexes, _ := config_.Get("root").StringList(false)

	return NewStatic(root, indexes), nil
}

// Handler interface
// HandlerFunc signature
func (self *Static) Handle(context *Context) bool {
	context.RequestContext.Request.Header.Del("__path")
	context.RequestContext.Request.Header.Add("__path", "/"+context.Path)
	self.RequestHandler(context.RequestContext)
	context.RequestContext.Request.Header.Del("__path")
	return context.RequestContext.Response.StatusCode() != fasthttp.StatusNotFound
}
