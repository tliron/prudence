package rest

import (
	"fmt"
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
	// TODO: support indexes
	return &Static{
		Root: root,
	}
}

// platform.CreateFunc signature
func CreateStatic(config ard.StringMap, context *js.Context) (interface{}, error) {
	config_ := ard.NewNode(config)

	root, _ := config_.Get("root").String()
	if rootUrl, err := context.Resolve(root, true); err == nil {
		if rootFileUrl, ok := rootUrl.(*urlpkg.FileURL); ok {
			root = rootFileUrl.Path
		} else {
			return nil, fmt.Errorf("Static \"root\" is not a file: %v", rootUrl)
		}
	} else {
		return nil, err
	}
	indexes := platform.AsStringList(config_.Get("indexes").Value)

	return NewStatic(root, indexes), nil
}

// Handler interface
// HandleFunc signature
func (self *Static) Handle(context *Context) bool {
	path := filepath.Join(self.Root, context.Path)
	http.ServeFile(NewResponseWriterWrapper(context), context.Request.Direct, path)
	//http.ServeFile(context.Response.Direct, context.Request.Direct, path)
	if context.Response.Status != http.StatusNotFound {
		context.Response.Bypass = true
		return true
	} else {
		return false
	}
}
