package rest

import (
	"github.com/tliron/kutil/ard"
	"github.com/tliron/kutil/logging"
	"github.com/tliron/prudence/platform"
)

func init() {
	platform.RegisterType("route", CreateRoute)
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

// CreateFunc signature
func CreateRoute(config ard.StringMap, getRelativeURL platform.GetRelativeURL) (interface{}, error) {
	var self Route

	config_ := ard.NewNode(config)
	self.Name, _ = config_.Get("name").String(true)
	paths := platform.AsStringList(config_.Get("paths").Data)
	var err error
	if self.PathTemplates, err = NewPathTemplates(paths...); err != nil {
		return nil, err
	}
	if handler := config_.Get("handler").Data; handler != nil {
		if self.Handler, err = GetHandleFunc(handler); err != nil {
			return nil, err
		}
	}

	return &self, nil
}

// Handler interface
// HandleFunc signature
func (self *Route) Handle(context *Context) bool {
	if matches := self.Match(context.Path); matches != nil {
		context = context.Copy()

		if self.Name != "" {
			context.Log = logging.NewSubLogger(context.Log, self.Name)
		}

		for key, value := range matches {
			switch key {
			case PATH_VARIABLE:
				context.Path = value
				context.Log.Debugf("set path = %s", value)

			default:
				context.Variables[key] = value
				context.Log.Debugf("set variable %s = %s", key, value)
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
