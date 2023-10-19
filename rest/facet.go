package rest

import (
	"github.com/tliron/commonjs-goja"
	"github.com/tliron/go-ard"
)

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

// ([platform.CreateFunc] signature)
func CreateFacet(jsContext *commonjs.Context, config ard.StringMap) (any, error) {
	config_ := ard.With(config).ConvertSimilar().NilMeansZero()

	var self Facet

	if route, err := CreateRoute(jsContext, config); err == nil {
		self.Route = route.(*Route)
	} else {
		return nil, err
	}

	self.Handler = self.Handle

	var err error
	if self.Representations, err = CreateRepresentations(config_.Get("representations").Value, jsContext); err != nil {
		return nil, err
	}

	return &self, nil
}

// ([Handler] interface, [HandleFunc] signature)
func (self *Facet) Handle(restContext *Context) (bool, error) {
	if representation, contentType, language, ok := self.Representations.NegotiateBest(restContext); ok {
		restContext = restContext.Clone()
		restContext.Response.ContentType = contentType
		restContext.Response.Language = language
		return representation.Handle(restContext)
	} else {
		return false, nil
	}
}
