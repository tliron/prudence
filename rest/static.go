package rest

import (
	"github.com/tliron/kutil/ard"
	urlpkg "github.com/tliron/kutil/url"
	"github.com/tliron/prudence/platform"
	"github.com/valyala/fasthttp"
)

func init() {
	platform.RegisterType("static", CreateStatic)
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
		Compress:           true,
		CompressBrotli:     true,
		PathRewrite: func(context *fasthttp.RequestCtx) []byte {
			path := context.Request.Header.Peek("__path")
			log.Debugf("path: %s", path)
			return path
		},
	}

	return &Static{
		RequestHandler: fs.NewRequestHandler(),
	}
}

// CreateFunc signature
func CreateStatic(config ard.StringMap, getRelativeURL platform.GetRelativeURL) (interface{}, error) {
	config_ := ard.NewNode(config)

	root, _ := config_.Get("root").String(false)
	if rootUrl, err := getRelativeURL(root); err == nil {
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
	context.context.Request.Header.Del("__path")
	context.context.Request.Header.Add("__path", "/"+context.Path)
	self.RequestHandler(context.context)
	context.context.Request.Header.Del("__path")
	return !NotFound(context.context)
}
