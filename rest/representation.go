package rest

import (
	"github.com/tliron/kutil/ard"
	"github.com/tliron/kutil/js"
	"github.com/tliron/prudence/platform"
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

	if hooks := node.Get("hooks").Data; hooks != nil {
		hooks_ := hooks.(map[string]*js.Hook)
		if construct, ok := hooks_["construct"]; ok {
			self.Construct = NewRepresentationFunc(construct)
		}
		if describe, ok := hooks_["describe"]; ok {
			self.Describe = NewRepresentationFunc(describe)
		}
		if present, ok := hooks_["present"]; ok {
			self.Present = NewRepresentationFunc(present)
		}
	}

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

func CreateRepresentations(config ard.Value) (Representations, error) {
	self := make(Representations)

	representations := platform.AsConfigList(config)
	for _, representation := range representations {
		representation_ := ard.NewNode(representation)
		representation__, _ := CreateRepresentation(representation_)
		contentTypes := platform.AsStringList(representation_.Get("contentTypes").Data)
		// TODO:
		//charSets := asStringList(representation_.Get("charSets").Data)
		//languages := asStringList(representation_.Get("languages").Data)

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
