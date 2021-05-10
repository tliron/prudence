package rest

import (
	"errors"
	"strings"

	"github.com/tliron/kutil/ard"
	"github.com/tliron/kutil/js"
	"github.com/tliron/prudence/js/common"
)

func init() {
	Register("facet", CreateFacet)
}

//
// Facet
//

type Facet struct {
	*Route

	Representers map[string]Representer
}

func NewFacet(name string, paths []string) *Facet {
	self := Facet{
		Route:        NewRoute(name, paths, nil),
		Representers: make(map[string]Representer),
	}
	self.Handler = self.Handle
	return &self
}

// CreateFunc signature
func CreateFacet(config ard.StringMap, getRelativeURL common.GetRelativeURL) (interface{}, error) {
	self := Facet{
		Representers: make(map[string]Representer),
	}

	route, _ := CreateRoute(config, getRelativeURL)
	self.Route = route.(*Route)
	if self.Handler != nil {
		return nil, errors.New("cannot set \"handler\" on facet")
	}
	self.Handler = self.Handle

	config_ := ard.NewNode(config)
	representations, _ := config_.Get("representations").List(true)
	for _, representation := range representations {
		representation_ := ard.NewNode(representation)
		contentTypes, _ := representation_.Get("contentTypes").StringList(true)
		representer := representation_.Get("representer").Data.(*js.Hook)

		representer_ := func(context *Context) {
			representer.Call(nil, context)
		}

		if len(contentTypes) == 0 {
			// Default representer
			self.SetRepresenter("", representer_)
		} else {
			for _, contentType := range contentTypes {
				self.SetRepresenter(contentType, representer_)
			}
		}
	}

	return &self, nil
}

func (self *Facet) SetRepresenter(contentType string, representer Representer) {
	self.Representers[contentType] = representer
}

func (self *Facet) FindRepresenter(context *Context) (Representer, string, bool) {
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Accept
	accept := strings.Split(string(context.RequestContext.Request.Header.Peek("Accept")), ",")
	context.Log.Infof("ACCEPT: %s", accept)

	// TODO: text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8

	for _, contentType := range accept {
		if representer, ok := self.Representers[contentType]; ok {
			return representer, contentType, true
		}
	}

	// Default representer
	representer, ok := self.Representers[""]
	return representer, "", ok
}

// Handler interface
// HandlerFunc signature
func (self *Facet) Handle(context *Context) bool {
	if context.RequestContext.IsGet() || context.RequestContext.IsHead() {
		if representer, contentType, ok := self.FindRepresenter(context); ok {
			context = context.Copy()
			context.ContentType = contentType
			return representer.Handle(context)
		}
	}

	return false
}
