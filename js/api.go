package js

import (
	"fmt"
	"strings"
	"sync"

	"github.com/dop251/goja"
	"github.com/tliron/kutil/ard"
	"github.com/tliron/kutil/js"
	"github.com/tliron/kutil/logging"
	urlpkg "github.com/tliron/kutil/url"
	"github.com/tliron/prudence/render"
	"github.com/tliron/prudence/rest"
)

//
// API
//

type API struct {
	Url             urlpkg.URL
	Log             logging.Logger
	DefaultNotFound rest.Handler
}

func NewAPI(url urlpkg.URL) *API {
	return &API{
		Url:             url,
		Log:             log,
		DefaultNotFound: rest.DefaultNotFound,
	}
}

func (self *API) newRuntime() *goja.Runtime {
	runtime := goja.New()
	runtime.SetFieldNameMapper(js.CamelCaseMapper)
	runtime.Set("prudence", self)
	return runtime
}

func (self *API) Create(config ard.StringMap) (interface{}, error) {
	return rest.Create(config, self.getRelativeURL)
}

func (self *API) Import(url string) (interface{}, error) {
	if url_, err := self.getRelativeURL(url); err == nil {
		_, value, err := self.run(url_)
		return value, err
	} else {
		return nil, err
	}
}

func (self *API) Hook(url string, name string) (*js.Hook, error) {
	if url_, err := self.getRelativeURL(url); err == nil {
		if runtime, err := self.cachedRun(url_); err == nil {
			if name == "" {
				name = "hook"
			}
			value := runtime.Get(name)
			if callable, ok := goja.AssertFunction(value); ok {
				return js.NewHook(callable, runtime), nil
			} else {
				return nil, fmt.Errorf("no \"%s\" function", name)
			}
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

func (self *API) Load(url string) (string, error) {
	if url_, err := self.getRelativeURL(url); err == nil {
		return urlpkg.ReadString(url_)
	} else {
		return "", err
	}
}

func (self *API) Render(content string, renderer string) (string, error) {
	return render.Render(content, renderer, self.getRelativeURL)
}

// hook.GetRelativeURL signature
func (self *API) getRelativeURL(url string) (urlpkg.URL, error) {
	urlContext := urlpkg.NewContext()
	defer urlContext.Release()

	var origins []urlpkg.URL
	if self.Url != nil {
		origins = []urlpkg.URL{self.Url.Origin()}
	}

	return urlpkg.NewValidURL(url, origins, urlContext)
}

func (self *API) run(url urlpkg.URL) (*goja.Runtime, interface{}, error) {
	if script, err := urlpkg.ReadString(url); err == nil {
		// JST
		if strings.HasSuffix(url.String(), ".jst") {
			if script, err = RenderJST(script); err != nil {
				return nil, nil, err
			}
		}

		if program, err := goja.Compile(url.String(), script, true); err == nil {
			runtime := NewAPI(url).newRuntime()
			if value, err := runtime.RunProgram(program); err == nil {
				return runtime, value.Export(), nil
			} else {
				return nil, nil, err
			}
		} else {
			return nil, nil, err
		}
	} else {
		return nil, nil, err
	}
}

var runtimeCache sync.Map

func (self *API) cachedRun(url urlpkg.URL) (*goja.Runtime, error) {
	key := url.Key()
	if runtime_, loaded := runtimeCache.Load(key); loaded {
		// In cache
		return runtime_.(*goja.Runtime), nil
	} else {
		if runtime, _, err := self.run(url); err == nil {
			if runtime_, loaded := runtimeCache.LoadOrStore(key, runtime); loaded {
				// In cache
				return runtime_.(*goja.Runtime), nil
			} else {
				return runtime, nil
			}
		} else {
			return nil, err
		}
	}
}
