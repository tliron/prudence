package rest

import (
	"github.com/tliron/kutil/ard"
	"github.com/tliron/kutil/logging"
	"github.com/tliron/prudence/js/common"
)

func init() {
	Register("router", CreateRouter)
}

//
// Router
//
// A handler that attempts to delegate to a list of handlers (in sequence)
//

type Router struct {
	Name     string
	Handlers []HandlerFunc
	Routes   []*Route
}

func NewRouter(name string) *Router {
	return &Router{
		Name: name,
	}
}

// CreateFunc signature
func CreateRouter(config ard.StringMap, getRelativeURL common.GetRelativeURL) (interface{}, error) {
	var self Router

	config_ := ard.NewNode(config)
	self.Name, _ = config_.Get("name").String(true)
	routes, _ := config_.Get("routes").List(true)
	for _, route := range routes {
		if route_, ok := route.(ard.StringMap); ok {
			route__, _ := CreateRoute(route_, getRelativeURL)
			route___ := route__.(*Route)
			self.Routes = append(self.Routes, route___)
			self.AddHandler(route___.Handle)
		}
	}

	return &self, nil
}

func (self *Router) AddHandler(handler HandlerFunc) {
	self.Handlers = append(self.Handlers, handler)
}

func (self *Router) AddRoute(route *Route) {
	self.Routes = append(self.Routes, route)
	self.AddHandler(route.Handle)
}

// Handler interface
// HandlerFunc signature
func (self *Router) Handle(context *Context) bool {
	if self.Name != "" {
		context = context.Copy()
		context.Log = logging.NewSubLogger(context.Log, self.Name)
	}

	for _, handler := range self.Handlers {
		if handled := handler(context); handled {
			return true
		}
	}

	return false
}
