package rest

import (
	"net/http"

	"github.com/tliron/commonjs-goja"
	"github.com/tliron/go-ard"
	"github.com/tliron/prudence/platform"
)

//
// Route
//
// Wraps a handler so that it would only be called if any
// single path template matches (in sequence)
//

type Route struct {
	Name                        string
	PathTemplates               PathTemplates
	RedirectTrailingSlashStatus int
	Variables                   map[string]any
	Handler                     HandleFunc
}

func NewRoute(name string) *Route {
	return &Route{
		Name:                        name,
		RedirectTrailingSlashStatus: http.StatusMovedPermanently, // 301
		Variables:                   make(map[string]any),
	}
}

// ([platform.CreateFunc] signature)
func CreateRoute(jsContext *commonjs.Context, config ard.StringMap) (any, error) {
	config_ := ard.With(config).ConvertSimilar().NilMeansZero()

	name, _ := config_.Get("name").String()

	self := NewRoute(name)

	var err error
	if paths := platform.AsStringList(config_.Get("paths")); len(paths) > 0 {
		if self.PathTemplates, err = NewPathTemplates(paths...); err != nil {
			return nil, err
		}
	}

	if redirectTrailingSlashStatus, ok := config_.Get("redirectTrailingSlashStatus").UnsignedInteger(); ok {
		self.RedirectTrailingSlashStatus = int(redirectTrailingSlashStatus)
	}

	if variables, ok := config_.Get("variables").StringMap(); ok {
		self.Variables = variables
	}

	if handler := config_.Get("handler"); handler != ard.NoNode {
		if self.Handler, err = GetHandleFunc(handler.Value, jsContext); err != nil {
			return nil, err
		}
	}

	return self, nil
}

// ([Handler] interface, [HandleFunc] signature)
func (self *Route) Handle(restContext *Context) (bool, error) {
	if matches := self.Match(restContext.Request.Path); matches != nil {
		restContext = restContext.AppendName(self.Name, true)

		ard.Merge(restContext.Variables, self.Variables, false)

		for key, value := range matches {
			switch key {
			case PathVariable:
				// Special handling for "*"
				restContext.Request.Path = value

			default:
				restContext.Variables[key] = value
			}
		}

		if self.Handler != nil {
			return self.Handler(restContext)
		}
	} else if self.PathTemplates.MatchAnyRedirectTrailingSlash(restContext.Request.Path) {
		return true, restContext.RedirectTrailingSlash(self.RedirectTrailingSlashStatus)
	}

	return false, nil
}

func (self *Route) Match(path string) map[string]string {
	if len(self.PathTemplates) == 0 {
		// Empty paths always matches
		return make(map[string]string)
	}

	return self.PathTemplates.MatchAny(path)
}
