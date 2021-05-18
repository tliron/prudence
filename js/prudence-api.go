package js

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/beevik/etree"
	"github.com/dop251/goja"
	"github.com/mitchellh/hashstructure/v2"
	"github.com/tliron/kutil/ard"
	formatpkg "github.com/tliron/kutil/format"
	"github.com/tliron/kutil/js"
	"github.com/tliron/kutil/logging"
	urlpkg "github.com/tliron/kutil/url"
	"github.com/tliron/kutil/util"
	"github.com/tliron/prudence/render"
	"github.com/tliron/prudence/rest"
	"github.com/tliron/yamlkeys"
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

func (self *API) Sprintf(format string, args ...interface{}) string {
	return fmt.Sprintf(format, args...)
}

func (self *API) JoinFilePath(elements ...string) string {
	return filepath.Join(elements...)
}

func (self *API) IsType(value ard.Value, type_ string) (bool, error) {
	// Special case whereby an integer stored as a float type has been optimized to an integer type
	if (type_ == "!!float") && ard.IsInteger(value) {
		return true, nil
	}

	if validate, ok := ard.TypeValidators[ard.TypeName(type_)]; ok {
		return validate(value), nil
	} else {
		return false, fmt.Errorf("unsupported type: %s", type_)
	}
}

func (self *API) ValidateFormat(code string, format string) error {
	return formatpkg.Validate(code, format)
}

func (self *API) Timestamp() ard.Value {
	return util.Timestamp(false)
}

func (self *API) NewXMLDocument() *etree.Document {
	return etree.NewDocument()
}

func (self *API) Decode(code string, format string, all bool) (ard.Value, error) {
	switch format {
	case "yaml", "":
		if all {
			if value, err := yamlkeys.DecodeAll(strings.NewReader(code)); err == nil {
				value_, _ := ard.MapsToStringMaps(value)
				return value_, err
			} else {
				return nil, err
			}
		} else {
			value, _, err := ard.DecodeYAML(code, false)
			value, _ = ard.MapsToStringMaps(value)
			return value, err
		}

	case "json":
		value, _, err := ard.DecodeJSON(code, false)
		value, _ = ard.MapsToStringMaps(value)
		return value, err

	case "cjson":
		value, _, err := ard.DecodeCompatibleJSON(code, false)
		value, _ = ard.MapsToStringMaps(value)
		return value, err

	case "xml":
		value, _, err := ard.DecodeCompatibleXML(code, false)
		value, _ = ard.MapsToStringMaps(value)
		return value, err

	case "cbor":
		value, _, err := ard.DecodeCBOR(code, false)
		value, _ = ard.MapsToStringMaps(value)
		return value, err

	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

func (self *API) Encode(value interface{}, format string, indent string, writer io.Writer) (string, error) {
	if writer == nil {
		return formatpkg.Encode(value, format, false)
	} else {
		err := formatpkg.Write(value, format, indent, false, writer)
		return "", err
	}
}

func (self *API) Exec(name string, arguments ...string) (string, error) {
	cmd := exec.Command(name, arguments...)
	if out, err := cmd.Output(); err == nil {
		return util.BytesToString(out), nil
	} else if err_, ok := err.(*exec.ExitError); ok {
		return "", fmt.Errorf("%s\n%s", err_.Error(), util.BytesToString(err_.Stderr))
	} else {
		return "", err
	}
}

func (self *API) TemporaryFile(pattern string, directory string) (string, error) {
	if file, err := os.CreateTemp(directory, pattern); err == nil {
		name := file.Name()
		os.Remove(name)
		return name, nil
	} else {
		return "", err
	}
}

func (self *API) TemporaryDirectory(pattern string, directory string) (string, error) {
	return os.MkdirTemp(directory, pattern)
}

func (self *API) Load(url string) (string, error) {
	if url_, err := self.GetRelativeURL(url); err == nil {
		return urlpkg.ReadString(url_)
	} else {
		return "", err
	}
}

func (self *API) Download(sourceUrl string, targetPath string) error {
	if sourceUrl_, err := self.GetRelativeURL(sourceUrl); err == nil {
		return urlpkg.DownloadTo(sourceUrl_, targetPath)
	} else {
		return err
	}
}

// Encode bytes as base64
func (self *API) Btoa(bytes []byte) string {
	return util.ToBase64(bytes)
}

// Decode base64 to bytes
func (self *API) Atob(b64 string) ([]byte, error) {
	// Note: if you need a string in JavaScript: String.fromCharCode.apply(null, .atob(...))
	return util.FromBase64(b64)
}

func (self *API) DeepCopy(value ard.Value) ard.Value {
	return ard.Copy(value)
}

func (self *API) DeepEquals(a ard.Value, b ard.Value) bool {
	return ard.Equals(a, b)
}

func (self *API) Hash(value interface{}) (string, error) {
	if hash, err := hashstructure.Hash(value, hashstructure.FormatV2, nil); err == nil {
		return strconv.FormatUint(hash, 10), nil
	} else {
		return "", err
	}
}

func (self *API) Create(config ard.StringMap) (interface{}, error) {
	return rest.Create(config, self.GetRelativeURL)
}

// Returns the last value
func (self *API) Import(url string) (interface{}, error) {
	if url_, err := self.GetRelativeURL(url); err == nil {
		_, value, err := self.run(url_)
		return value, err
	} else {
		return nil, err
	}
}

// Returns the global object
func (self *API) Require(url string) (interface{}, error) {
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

// Returns a hook to one function
func (self *API) Hook(url string, name string) (*js.Hook, error) {
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

// Returns hooks for all functions in the global object
func (self *API) Hooks(url string) (map[string]*js.Hook, error) {
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

func (self *API) Render(content string, renderer string) (string, error) {
	return render.Render(content, renderer, self.GetRelativeURL)
}

func (self *API) Start(startables interface{}) error {
	var startables_ ard.List
	var ok bool
	if startables_, ok = startables.(ard.List); !ok {
		startables_ = ard.List{startables}
	}

	for _, startable := range startables_ {
		if startable_, ok := startable.(rest.Startable); ok {
			go func() {
				if err := startable_.Start(); err != nil {
					self.Log.Errorf("%s", err.Error())
				}
			}()
		} else {
			return fmt.Errorf("object not startable: %T", startable)
		}
	}

	// Block forever
	<-make(chan struct{})

	return nil
}

// common.HasGetRelativeURL interface
// common.GetRelativeURL signature
func (self *API) GetRelativeURL(url string) (urlpkg.URL, error) {
	urlContext := urlpkg.NewContext()
	defer urlContext.Release()

	var origins []urlpkg.URL
	if self.Url != nil {
		origins = []urlpkg.URL{self.Url.Origin()}
	}

	return urlpkg.NewValidURL(url, origins, urlContext)
}

//var registry = require.NewRegistry(require.WithGlobalFolders("npm"))

func (self *API) newRuntime() *goja.Runtime {
	runtime := goja.New()
	//registry.Enable(runtime)
	runtime.SetFieldNameMapper(js.CamelCaseMapper)
	runtime.Set("prudence", self)
	runtime.Set("require", self.Require)
	return runtime
}

func (self *API) run(url urlpkg.URL) (*goja.Runtime, interface{}, error) {
	if script, err := urlpkg.ReadString(url); err == nil {
		if strings.HasSuffix(url.String(), ".jst") {
			// JST
			if script, err = RenderJST(script); err != nil {
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
