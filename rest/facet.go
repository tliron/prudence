package rest

import (
	"errors"

	"github.com/tliron/kutil/ard"
	"github.com/tliron/kutil/js"
	"github.com/tliron/prudence/platform"
)

func init() {
	platform.RegisterType("Facet", CreateFacet)
}

//
// Facet
//

type Facet struct {
	*Route

	Representations *Representations
}

func NewFacet(name string) *Facet {
	self := Facet{
		Route:           NewRoute(name),
		Representations: new(Representations),
	}
	self.Handler = self.Handle
	return &self
}

// CreateFunc signature
func CreateFacet(config ard.StringMap, context *js.Context) (interface{}, error) {
	var self Facet

	if route, err := CreateRoute(config, context); err == nil {
		self.Route = route.(*Route)
	} else {
		return nil, err
	}
	if self.Handler != nil {
		return nil, errors.New("cannot set \"handler\" on facet")
	}
	self.Handler = self.Handle

	config_ := ard.NewNode(config)
	var err error
	if self.Representations, err = CreateRepresentations(config_.Get("representations").Data, context); err != nil {
		return nil, err
	}

	return &self, nil
}

// Handler interface
// HandleFunc signature
func (self *Facet) Handle(context *Context) bool {
	if representation, contentType, ok := self.Representations.NegotiateBest(context); ok {
		context = context.Copy()
		context.Response.ContentType = contentType
		return representation.Handle(context)
	} else {
		return false
	}
}
