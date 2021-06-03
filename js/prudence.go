package js

import (
	"fmt"
	"html"

	"github.com/tliron/kutil/ard"
	"github.com/tliron/kutil/js"
	"github.com/tliron/kutil/logging"
	urlpkg "github.com/tliron/kutil/url"
	platform "github.com/tliron/prudence/platform"
	rest "github.com/tliron/prudence/rest"
)

//
// PrudenceAPI
//

type PrudenceAPI struct {
	js.UtilAPI
	js.FormatAPI
	js.FileAPI

	Log             logging.Logger
	JsContext       *js.Context
	DefaultNotFound rest.Handler
}

func NewPrudenceAPI(urlContext *urlpkg.Context, jsContext *js.Context) *PrudenceAPI {
	return &PrudenceAPI{
		FileAPI:         js.NewFileAPI(urlContext),
		Log:             log,
		JsContext:       jsContext,
		DefaultNotFound: rest.DefaultNotFound,
	}
}

func (self *PrudenceAPI) LoadString(id string) (string, error) {
	if url_, err := self.JsContext.Resolve(id); err == nil {
		if fileUrl, ok := url_.(*urlpkg.FileURL); ok {
			if self.JsContext.Environment.Watcher != nil {
				self.JsContext.Environment.Watcher.Add(fileUrl.Path)
			}
		}

		return urlpkg.ReadString(url_)
	} else {
		return "", err
	}
}

func (self *PrudenceAPI) EscapeHtml(text string) string {
	return html.EscapeString(text)
}

func (self *PrudenceAPI) UnescapeHtml(text string) string {
	return html.UnescapeString(text)
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
	return platform.Render(content, renderer, self.JsContext)
}
