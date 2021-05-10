package rest

import (
	"github.com/tliron/kutil/ard"
	"github.com/tliron/kutil/logging"
	"github.com/tliron/prudence/js/common"
)

func init() {
	Register("route", CreateRoute)
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
	Handler       HandlerFunc
}

func NewRoute(name string, paths []string, handler HandlerFunc) *Route {
	return &Route{
		Name:          name,
		PathTemplates: NewPathTemplates(paths),
		Handler:       handler,
	}
}

// CreateFunc signature
func CreateRoute(config ard.StringMap, getRelativeURL common.GetRelativeURL) (interface{}, error) {
	var self Route

	config_ := ard.NewNode(config)
	self.Name, _ = config_.Get("name").String(true)
	paths, _ := config_.Get("paths").StringList(true)
	self.PathTemplates = NewPathTemplates(paths)
	handler := config_.Get("handler").Data
	self.Handler, _ = GetHandler(handler)

	return &self, nil
}

// Handler interface
// HandlerFunc signature
func (self *Route) Handle(context *Context) bool {
	if matches := self.Match(context.Path); matches != nil {
		context = context.Copy()

		if self.Name != "" {
			context.Log = logging.NewSubLogger(context.Log, self.Name)
		}

		for key, value := range matches {
			switch key {
			case "PATH":
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
