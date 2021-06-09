package rest

import (
	"github.com/tliron/kutil/ard"
	"github.com/tliron/kutil/js"
	"github.com/tliron/prudence/platform"
)

func init() {
	platform.RegisterType("Router", CreateRouter)
}

//
// Router
//
// A handler that attempts to delegate to a list of handlers (in sequence)
//

type Router struct {
	Name     string
	Handlers []HandleFunc
	Routes   []*Route
}

func NewRouter(name string) *Router {
	return &Router{
		Name: name,
	}
}

// CreateFunc signature
func CreateRouter(config ard.StringMap, context *js.Context) (interface{}, error) {
	var self Router

	config_ := ard.NewNode(config)
	self.Name, _ = config_.Get("name").String(true)
	routes := platform.AsConfigList(config_.Get("routes").Data)
	for _, route := range routes {
		if route_, ok := route.(ard.StringMap); ok {
			if route__, err := CreateRoute(route_, context); err == nil {
				route___ := route__.(*Route)
				self.Routes = append(self.Routes, route___)
				self.AddHandler(route___.Handle)
			} else {
				return nil, err
			}
		}
	}

	return &self, nil
}

func (self *Router) AddHandler(handler HandleFunc) {
	self.Handlers = append(self.Handlers, handler)
}

func (self *Router) AddRoute(route *Route) {
	self.Routes = append(self.Routes, route)
	self.AddHandler(route.Handle)
}

// Handler interface
// HandleFunc signature
func (self *Router) Handle(context *Context) bool {
	context = context.AddName(self.Name)

	for _, handler := range self.Handlers {
		if handled := handler(context); handled {
			return true
		}
	}

	return false
}
