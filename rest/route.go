package rest

import (
	"github.com/tliron/commonjs-goja"
	"github.com/tliron/go-ard"
	"github.com/tliron/prudence/platform"
)

func init() {
	platform.RegisterType("Route", CreateRoute)
}

//
// Route
//
// Wraps a handler so that it would only be called if any
// single path template matches (in sequence)
//

type Route struct {
	Name          string
	PathTemplates PathTemplates
	Handler       HandleFunc
}

func NewRoute(name string) *Route {
	return &Route{
		Name: name,
	}
}

// platform.CreateFunc signature
func CreateRoute(config ard.StringMap, context *commonjs.Context) (interface{}, error) {
	var self Route

	config_ := ard.NewNode(config)
	self.Name, _ = config_.Get("name").NilMeansZero().String()
	paths := platform.AsStringList(config_.Get("paths").Value)
	var err error
	if self.PathTemplates, err = NewPathTemplates(paths...); err != nil {
		return nil, err
	}
	if handler := config_.Get("handler").Value; handler != nil {
		if self.Handler, err = GetHandleFunc(handler, context); err != nil {
			return nil, err
		}
	}

	return &self, nil
}

// Handler interface
// HandleFunc signature
func (self *Route) Handle(context *Context) bool {
	if matches := self.Match(context.Path); matches != nil {
		if context_ := context.AppendName(self.Name); context == context_ {
			context = context.Copy()
		} else {
			context = context_
		}

		for key, value := range matches {
			switch key {
			case PathVariable:
				context.Path = value

			default:
				context.Variables[key] = value
			}
		}

		if self.Handler != nil {
			return self.Handler(context)
		}
	}

	return false
}

func (self *Route) Match(path string) map[string]string {
	if len(self.PathTemplates) == 0 {
		// Empty paths always matches
		return make(map[string]string)
	}

	return self.PathTemplates.MatchAny(path)
}
