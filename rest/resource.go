package rest

import (
	"github.com/tliron/kutil/ard"
	"github.com/tliron/kutil/js"
	"github.com/tliron/prudence/platform"
)

func init() {
	platform.RegisterType("Resource", CreateResource)
}

//
// Resource
//

type Resource struct {
	*Router

	Facets []*Facet
}

func NewResource(name string) *Resource {
	return &Resource{
		Router: NewRouter(name),
	}
}

// CreateFunc signature
func CreateResource(config ard.StringMap, context *js.Context) (interface{}, error) {
	var self Resource

	router, _ := CreateRouter(config, context)
	self.Router = router.(*Router)

	config_ := ard.NewNode(config)
	facets := platform.AsConfigList(config_.Get("facets").Data)
	for _, facet := range facets {
		if facet_, ok := facet.(ard.StringMap); ok {
			if facet__, err := CreateFacet(facet_, context); err == nil {
				self.AddFacet(facet__.(*Facet))
			} else {
				return nil, err
			}
		}
	}

	return &self, nil
}

func (self *Resource) AddFacet(facet *Facet) {
	self.Facets = append(self.Facets, facet)
	self.AddRoute(facet.Route)
}
