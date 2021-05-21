package rest

import (
	"errors"

	"github.com/tliron/kutil/ard"
	"github.com/tliron/prudence/platform"
)

func init() {
	platform.RegisterType("facet", CreateFacet)
}

//
// Facet
//

type Facet struct {
	*Route

	Representations Representations
}

func NewFacet(name string, paths []string) *Facet {
	self := Facet{
		Route:           NewRoute(name, paths, nil),
		Representations: make(Representations),
	}
	self.Handler = self.Handle
	return &self
}

// CreateFunc signature
func CreateFacet(config ard.StringMap, getRelativeURL platform.GetRelativeURL) (interface{}, error) {
	self := Facet{
		Representations: make(Representations),
	}

	route, _ := CreateRoute(config, getRelativeURL)
	self.Route = route.(*Route)
	if self.Handler != nil {
		return nil, errors.New("cannot set \"handler\" on facet")
	}
	self.Handler = self.Handle

	config_ := ard.NewNode(config)
	self.Representations, _ = CreateRepresentations(config_.Get("representations").Data)

	return &self, nil
}

func (self *Facet) FindRepresentation(context *Context) (*Representation, string, bool) {
	for _, contentType := range ParseAccept(context) {
		if functions, ok := self.Representations[contentType]; ok {
			return functions, contentType, true
		}
	}

	// Default representation
	functions, ok := self.Representations[""]
	return functions, "", ok
}

// Handler interface
// HandleFunc signature
func (self *Facet) Handle(context *Context) bool {
	if representation, contentType, ok := self.FindRepresentation(context); ok {
		context = context.Copy()
		context.ContentType = contentType
		return representation.Handle(context)
	} else {
		return false
	}
}
