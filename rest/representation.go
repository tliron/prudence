package rest

import (
	"github.com/tliron/kutil/ard"
	"github.com/tliron/kutil/js"
)

//
// RepresentionFunc
//

type RepresentionFunc func(context *Context) error

func NewRepresentationFunc(hook *js.Hook) RepresentionFunc {
	if hook != nil {
		return func(context *Context) error {
			_, err := hook.Call(nil, context)
			return err
		}
	} else {
		return nil
	}
}

//
// Represention
//

type Representation struct {
	Construct RepresentionFunc
	Describe  RepresentionFunc
	Present   RepresentionFunc
}

func CreateRepresentation(node *ard.Node) (*Representation, error) {
	var self Representation
	if construct := node.Get("construct").Data; construct != nil {
		self.Construct = NewRepresentationFunc(construct.(*js.Hook))
	}
	if describe := node.Get("describe").Data; describe != nil {
		self.Describe = NewRepresentationFunc(describe.(*js.Hook))
	}
	if present := node.Get("present").Data; present != nil {
		self.Present = NewRepresentationFunc(present.(*js.Hook))
	}
	return &self, nil
}

//
// Representations
//

type Representations map[string]*Representation

func CreateRepresentations(configs ard.List) (Representations, error) {
	self := make(Representations)

	for _, representation := range configs {
		representation_ := ard.NewNode(representation)
		representation__, _ := CreateRepresentation(representation_)

		contentTypes, _ := representation_.Get("contentTypes").StringList(true)
		if len(contentTypes) == 0 {
			// Default representation
			self[""] = representation__
		} else {
			for _, contentType := range contentTypes {
				self[contentType] = representation__
			}
		}
	}

	return self, nil
}
