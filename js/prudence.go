package js

import (
	"fmt"

	"github.com/tliron/kutil/ard"
	"github.com/tliron/kutil/js"
	"github.com/tliron/kutil/logging"
	urlpkg "github.com/tliron/kutil/url"
	"github.com/tliron/prudence/platform"
	"github.com/tliron/prudence/rest"
)

//
// PrudenceAPI
//

type PrudenceAPI struct {
	js.UtilAPI
	js.FormatAPI
	js.FileAPI

	Log             logging.Logger
	JSContext       *js.Context
	DefaultNotFound rest.Handler
}

func NewPrudenceAPI(urlContext *urlpkg.Context, jsContext *js.Context) *PrudenceAPI {
	return &PrudenceAPI{
		FileAPI:         js.NewFileAPI(urlContext),
		Log:             log,
		JSContext:       jsContext,
		DefaultNotFound: rest.DefaultNotFound,
	}
}

func (self *PrudenceAPI) LoadString(id string) (string, error) {
	if url_, err := self.JSContext.Resolve(id); err == nil {
		if fileUrl, ok := url_.(*urlpkg.FileURL); ok {
			if self.JSContext.Environment.Watcher != nil {
				self.JSContext.Environment.Watcher.Add(fileUrl.Path)
			}
		}

		return urlpkg.ReadString(url_)
	} else {
		return "", err
	}
}

// Platform

func (self *PrudenceAPI) Start(startables interface{}) error {
	var list []ard.Value
	var ok bool
	if list, ok = startables.(ard.List); !ok {
		list = ard.List{startables}
	}

	var startables_ []platform.Startable

	for _, startable := range list {
		if startable_, ok := startable.(platform.Startable); ok {
			startables_ = append(startables_, startable_)
		} else {
			return fmt.Errorf("object not startable: %T", startable)
		}
	}

	return platform.Start(startables_)
}

func (self *PrudenceAPI) SetCache(cacheBackend platform.CacheBackend) {
	platform.SetCacheBackend(cacheBackend)
}

func (self *PrudenceAPI) InvalidateCacheGroup(group string) {
	if cacheBackend := platform.GetCacheBackend(); cacheBackend != nil {
		cacheBackend.DeleteGroup(platform.CacheKey(group))
	}
}

func (self *PrudenceAPI) Render(content string, renderer string) (string, error) {
	return platform.Render(content, renderer, self.JSContext.Resolve)
}
