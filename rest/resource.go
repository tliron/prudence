package rest

import (
	"github.com/tliron/kutil/ard"
	"github.com/tliron/prudence/js/common"
)

func init() {
	Register("resource", CreateResource)
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
func CreateResource(config ard.StringMap, getRelativeURL common.GetRelativeURL) (interface{}, error) {
	var self Resource

	router, _ := CreateRouter(config, getRelativeURL)
	self.Router = router.(*Router)

	config_ := ard.NewNode(config)
	facets, _ := config_.Get("facets").List(true)
	for _, facet := range facets {
		if facet_, ok := facet.(ard.StringMap); ok {
			facet__, _ := CreateFacet(facet_, getRelativeURL)
			self.AddFacet(facet__.(*Facet))
		}
	}

	return &self, nil
}

func (self *Resource) AddFacet(facet *Facet) {
	self.Facets = append(self.Facets, facet)
	self.AddRoute(facet.Route)
}
