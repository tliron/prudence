package rest

import (
	"errors"
	"strings"

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

	Representations Representations
}

func NewFacet(name string) *Facet {
	self := Facet{
		Route:           NewRoute(name),
		Representations: make(Representations),
	}
	self.Handler = self.Handle
	return &self
}

// CreateFunc signature
func CreateFacet(config ard.StringMap, context *js.Context) (interface{}, error) {
	self := Facet{
		Representations: make(Representations),
	}

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
	if self.Representations, err = CreateRepresentations(config_.Get("representations").Data, context.Environment.Runtime); err != nil {
		return nil, err
	}

	return &self, nil
}

func (self *Facet) NegotiateBestRepresentation(context *Context) (*Representation, string, bool) {
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Accept
	// Example: text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8
	//TODO: sorted by preference
	clientContentTypes := strings.Split(context.Request.Header.Get(HeaderAccept), ",")
	for _, clientContentType := range clientContentTypes {
		for serverContentType, functions := range self.Representations {
			if clientContentType == serverContentType {
				return functions, serverContentType, true
			}
		}
	}

	// Default representation
	functions, ok := self.Representations[""]
	return functions, "", ok
}

// Handler interface
// HandleFunc signature
func (self *Facet) Handle(context *Context) bool {
	if representation, contentType, ok := self.NegotiateBestRepresentation(context); ok {
		context = context.Copy()
		context.Response.ContentType = contentType
		return representation.Handle(context)
	} else {
		return false
	}
}
