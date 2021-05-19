package js

import (
	"fmt"
	"strings"
	"sync"

	"github.com/dop251/goja"
	"github.com/tliron/kutil/js"
	urlpkg "github.com/tliron/kutil/url"
	"github.com/tliron/prudence/platform"
)

// Returns the global object (cached)
func (self *PrudenceAPI) Require(url string) (interface{}, error) {
	if url_, err := self.GetRelativeURL(url); err == nil {
		if runtime, err := self.cachedRun(url_); err == nil {
			return runtime.GlobalObject().Export(), nil
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

// Returns the last value (not cached)
func (self *PrudenceAPI) Run(url string) (interface{}, error) {
	if url_, err := self.GetRelativeURL(url); err == nil {
		_, value, err := self.run(url_)
		return value, err
	} else {
		return nil, err
	}
}

// Returns a hook to one function (cached)
func (self *PrudenceAPI) Hook(url string, name string) (*js.Hook, error) {
	if url_, err := self.GetRelativeURL(url); err == nil {
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

// Returns hooks for all functions in the global object (cached)
func (self *PrudenceAPI) Hooks(url string) (map[string]*js.Hook, error) {
	if url_, err := self.GetRelativeURL(url); err == nil {
		if runtime, err := self.cachedRun(url_); err == nil {
			hooks := make(map[string]*js.Hook)
			global := runtime.GlobalObject()
			for _, key := range global.Keys() {
				value := global.Get(key)
				if callable, ok := goja.AssertFunction(value); ok {
					hooks[key] = js.NewHook(callable, runtime)
				}
			}
			return hooks, nil
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

//var registry = require.NewRegistry(require.WithGlobalFolders("npm"))

func (self *PrudenceAPI) newRuntime() *goja.Runtime {
	runtime := goja.New()
	//registry.Enable(runtime)
	runtime.SetFieldNameMapper(js.CamelCaseMapper)

	runtime.Set("prudence", self)
	runtime.Set("require", self.Require)

	platform.OnAPIs(func(name string, api interface{}) bool {
		runtime.Set(name, api)
		return true
	})

	return runtime
}

func (self *PrudenceAPI) run(url urlpkg.URL) (*goja.Runtime, interface{}, error) {
	if script, err := urlpkg.ReadString(url); err == nil {
		if strings.HasSuffix(url.String(), ".jst") {
			// JST
			if script, err = platform.Render(script, "jst", self.GetRelativeURL); err != nil {
				return nil, nil, err
			}
		}

		/*else if strings.HasSuffix(url.String(), ".ts") {
			// TypeScript
			if script, err = render.RenderTypeScript(script, self.GetRelativeURL); err != nil {
				return nil, nil, err
			}
			self.Log.Debug(script)
		} else if strings.HasSuffix(url.String(), ".tsx") {
			// TSX
			if script, err = render.RenderTSX(script, self.GetRelativeURL); err != nil {
				return nil, nil, err
			}
			self.Log.Debug(script)
		} else if strings.HasSuffix(url.String(), ".jsx") {
			// JSX
			if script, err = render.RenderJSX(script, self.GetRelativeURL); err != nil {
				return nil, nil, err
			}
			self.Log.Debug(script)
		}*/

		if program, err := goja.Compile(url.String(), script, true); err == nil {
			runtime := NewPrudenceAPI(url).newRuntime()
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

func (self *PrudenceAPI) cachedRun(url urlpkg.URL) (*goja.Runtime, error) {
	key := url.Key()
	// TODO: global lock!
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
