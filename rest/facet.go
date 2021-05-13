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

	Describers map[string]RepresentionFunc
	Presenters map[string]RepresentionFunc
}

func NewFacet(name string, paths []string) *Facet {
	self := Facet{
		Route:      NewRoute(name, paths, nil),
		Describers: make(map[string]RepresentionFunc),
		Presenters: make(map[string]RepresentionFunc),
	}
	self.Handler = self.Handle
	return &self
}

// CreateFunc signature
func CreateFacet(config ard.StringMap, getRelativeURL common.GetRelativeURL) (interface{}, error) {
	self := Facet{
		Describers: make(map[string]RepresentionFunc),
		Presenters: make(map[string]RepresentionFunc),
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
		var describer RepresentionFunc
		if describer_ := representation_.Get("describer").Data; describer_ != nil {
			describer = NewRepresentationFunc(describer_.(*js.Hook))
		}
		var presenter RepresentionFunc
		if presenter_ := representation_.Get("presenter").Data; presenter_ != nil {
			presenter = NewRepresentationFunc(presenter_.(*js.Hook))
		}

		if len(contentTypes) == 0 {
			// Defaults
			if describer != nil {
				self.SetDescriber("", describer)
			}
			if presenter != nil {
				self.SetPresenter("", presenter)
			}
		} else {
			for _, contentType := range contentTypes {
				if describer != nil {
					self.SetDescriber(contentType, describer)
				}
				if presenter != nil {
					self.SetPresenter(contentType, presenter)
				}
			}
		}
	}

	return &self, nil
}

func (self *Facet) SetDescriber(contentType string, describer RepresentionFunc) {
	self.Describers[contentType] = describer
}

func (self *Facet) SetPresenter(contentType string, presenter RepresentionFunc) {
	self.Presenters[contentType] = presenter
}

func (self *Facet) FindDescriber(context *Context) (RepresentionFunc, string, bool) {
	for _, contentType := range parseAccept(context) {
		if describer, ok := self.Describers[contentType]; ok {
			return describer, contentType, true
		}
	}

	// Default describer
	describer, ok := self.Describers[""]
	return describer, "", ok
}

func (self *Facet) FindPresenter(context *Context) (RepresentionFunc, string, bool) {
	for _, contentType := range parseAccept(context) {
		if presenter, ok := self.Presenters[contentType]; ok {
			return presenter, contentType, true
		}
	}

	// Default presenter
	presenter, ok := self.Presenters[""]
	return presenter, "", ok
}

// Handler interface
// HandleFunc signature
func (self *Facet) Handle(context *Context) bool {
	context = context.Copy()
	context.CacheKey = context.context.URI().String()

	if context.context.IsHead() {
		if describer, contentType, ok := self.FindDescriber(context); ok {
			context.ContentType = contentType
			describer.Call(context)
			return true
		}
	}

	if presenter, contentType, ok := self.FindPresenter(context); ok {
		context.ContentType = contentType
		presenter.Call(context)
		return true
	}

	return false
}

func parseAccept(context *Context) []string {
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Accept
	// TODO: text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8
	accept := strings.Split(string(context.context.Request.Header.Peek("Accept")), ",")
	context.Log.Infof("ACCEPT: %s", accept)
	return accept
}
