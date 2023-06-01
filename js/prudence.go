package js

import (
	contextpkg "context"
	"errors"
	"fmt"
	"html"
	"time"

	"github.com/tliron/commonjs-goja"
	"github.com/tliron/commonlog"
	"github.com/tliron/exturl"
	"github.com/tliron/go-ard"
	"github.com/tliron/kutil/util"
	"github.com/tliron/prudence/platform"
	"github.com/tliron/prudence/rest"
)

const DEFAULT_TIMEOUT_SECONDS = 10.0

//
// PrudenceAPI
//

type PrudenceAPI struct {
	commonjs.UtilAPI
	commonjs.TranscribeAPI
	commonjs.FileAPI

	Arguments       map[string]string
	Log             commonlog.Logger
	JsContext       *commonjs.Context
	DefaultNotFound rest.Handler
}

func NewPrudenceAPI(urlContext *exturl.Context, jsContext *commonjs.Context, arguments map[string]string) *PrudenceAPI {
	return &PrudenceAPI{
		FileAPI:         commonjs.NewFileAPI(urlContext),
		Arguments:       arguments,
		Log:             log,
		JsContext:       jsContext,
		DefaultNotFound: rest.DefaultNotFound,
	}
}

func (self *PrudenceAPI) LoadString(context contextpkg.Context, id string) (string, error) {
	if bytes, err := self.LoadBytes(context, id); err == nil {
		return util.BytesToString(bytes), nil
	} else {
		return "", err
	}
}

func (self *PrudenceAPI) LoadBytes(context contextpkg.Context, id string) ([]byte, error) {
	if url_, err := self.JsContext.Resolve(id, true); err == nil {
		if fileUrl, ok := url_.(*exturl.FileURL); ok {
			if err := self.JsContext.Environment.Watch(fileUrl.Path); err != nil {
				return nil, err
			}
		}

		return exturl.ReadBytes(context, url_)
	} else {
		return nil, err
	}
}

func (self *PrudenceAPI) EscapeHtml(text string) string {
	return html.EscapeString(text)
}

func (self *PrudenceAPI) UnescapeHtml(text string) string {
	return html.UnescapeString(text)
}

// Platform

func (self *PrudenceAPI) Start(startables interface{}, timeoutSeconds float64) error {
	var list []ard.Value
	var ok bool
	if list, ok = startables.(ard.List); !ok {
		list = ard.List{startables}
	}

	var startables_ []platform.Startable

	add := func(o interface{}) bool {
		added := false
		if hasStartables, ok := o.(platform.HasStartables); ok {
			startables_ = append(startables_, hasStartables.GetStartables()...)
			added = true
		}
		if startable, ok := o.(platform.Startable); ok {
			startables_ = append(startables_, startable)
			added = true
		}
		return added
	}

	add(platform.GetCacheBackend())
	add(platform.GetScheduler())

	for _, startable := range list {
		if !add(startable) {
			return fmt.Errorf("object not startable: %T", startable)
		}
	}

	if timeoutSeconds == 0.0 {
		timeoutSeconds = DEFAULT_TIMEOUT_SECONDS
	}

	return platform.Start(startables_, time.Duration(timeoutSeconds*float64(time.Second)))
}

func (self *PrudenceAPI) SetCache(cacheBackend platform.CacheBackend) {
	platform.SetCacheBackend(cacheBackend)
}

func (self *PrudenceAPI) InvalidateCacheGroup(group string) {
	if cacheBackend := platform.GetCacheBackend(); cacheBackend != nil {
		cacheBackend.DeleteGroup(platform.CacheKey(group))
	}
}

func (self *PrudenceAPI) SetScheduler(scheduler platform.Scheduler) {
	platform.SetScheduler(scheduler)
}

func (self *PrudenceAPI) Schedule(cronPattern string, job func()) error {
	if scheduler := platform.GetScheduler(); scheduler != nil {
		return scheduler.Schedule(cronPattern, job)
	} else {
		return errors.New("no scheduler")
	}
}

func (self *PrudenceAPI) Render(content string, renderer string) (string, error) {
	return platform.Render(content, renderer, self.JsContext)
}
