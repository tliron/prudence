package rest

import (
	"github.com/tliron/commonjs-goja"
	"github.com/tliron/go-ard"
	"github.com/tliron/prudence/platform"
)

//
// Router
//
// A handler that attempts to delegate to a list of handlers (in sequence)
//

type Router struct {
	Name      string
	Variables map[string]any
	Handlers  []HandleFunc
	Routes    []*Route
}

func NewRouter(name string) *Router {
	return &Router{
		Name:      name,
		Variables: make(map[string]any),
	}
}

// ([platform.CreateFunc] signature)
func CreateRouter(jsContext *commonjs.Context, config ard.StringMap) (any, error) {
	config_ := ard.With(config).ConvertSimilar().NilMeansZero()

	name, _ := config_.Get("name").String()

	self := NewRouter(name)

	if variables, ok := config_.Get("variables").StringMap(); ok {
		self.Variables = variables
	}

	if err := platform.CreateFromConfigList(jsContext, config_.Get("routes").Value, "Route", func(instance any, config__ ard.StringMap) {
		route := instance.(*Route)
		self.Routes = append(self.Routes, route)
		self.AddHandler(route.Handle)
	}); err != nil {
		return nil, err
	}

	return self, nil
}

func (self *Router) AddHandler(handler HandleFunc) {
	self.Handlers = append(self.Handlers, handler)
}

func (self *Router) AddRoute(route *Route) {
	self.Routes = append(self.Routes, route)
	self.AddHandler(route.Handle)
}

// ([Handler] interface, [HandleFunc] signature)
func (self *Router) Handle(restContext *Context) (bool, error) {
	restContext = restContext.AppendName(self.Name, false)

	ard.Merge(restContext.Variables, self.Variables, false)

	for _, handler := range self.Handlers {
		if handled, err := handler(restContext); err == nil {
			if handled {
				return true, nil
			}
		} else {
			return false, err
		}
	}

	return false, nil
}
