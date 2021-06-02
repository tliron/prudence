package rest

import (
	"github.com/tliron/kutil/ard"
	"github.com/tliron/kutil/js"
	urlpkg "github.com/tliron/kutil/url"
	"github.com/tliron/kutil/util"
	"github.com/tliron/prudence/platform"
	"github.com/valyala/fasthttp"
)

func init() {
	platform.RegisterType("Static", CreateStatic)
}

//
// Static
//

type Static struct {
	RequestHandler fasthttp.RequestHandler

	cleanStop chan struct{}
}

func NewStatic(root string, indexes []string) *Static {
	cleanStop := make(chan struct{})

	fs := fasthttp.FS{
		Root:               root,
		IndexNames:         indexes,
		GenerateIndexPages: len(indexes) > 0,
		Compress:           true,
		CompressBrotli:     true,
		CleanStop:          cleanStop,
		PathRewrite: func(context *fasthttp.RequestCtx) []byte {
			path := context.Request.Header.Peek(PATH_HEADER)
			log.Debugf("path: %s", path)
			return path
		},
	}

	self := Static{
		RequestHandler: fs.NewRequestHandler(),
		cleanStop:      cleanStop,
	}

	util.OnExit(self.Close)

	return &self
}

// CreateFunc signature
func CreateStatic(config ard.StringMap, context *js.Context) (interface{}, error) {
	config_ := ard.NewNode(config)

	root, _ := config_.Get("root").String(false)
	if rootUrl, err := context.Resolve(root); err == nil {
		root = rootUrl.(*urlpkg.FileURL).Path
	} else {
		return nil, err
	}
	indexes := platform.AsStringList(config_.Get("indexes").Data)

	return NewStatic(root, indexes), nil
}

func (self *Static) Close() {
	close(self.cleanStop)
}

// Handler interface
// HandleFunc signature
func (self *Static) Handle(context *Context) bool {
	context.Context.Request.Header.Del(PATH_HEADER)
	context.Context.Request.Header.Add(PATH_HEADER, "/"+context.Path)
	self.RequestHandler(context.Context)
	context.Context.Request.Header.Del(PATH_HEADER)
	return !NotFound(context.Context)
}
