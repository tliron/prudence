package rest

import (
	"github.com/tliron/commonjs-goja"
	"github.com/tliron/go-ard"
	"github.com/tliron/prudence/platform"
)

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

// ([platform.CreateFunc] signature)
func CreateResource(jsContext *commonjs.Context, config ard.StringMap) (any, error) {
	config_ := ard.With(config).ConvertSimilar().NilMeansZero()

	var self Resource

	if router, err := CreateRouter(jsContext, config); err == nil {
		self.Router = router.(*Router)
	} else {
		return nil, err
	}

	if err := platform.CreateFromConfigList(jsContext, config_.Get("facets").Value, "Facet", func(instance any, config__ ard.StringMap) {
		self.AddFacet(instance.(*Facet))
	}); err != nil {
		return nil, err
	}

	return &self, nil
}

func (self *Resource) AddFacet(facet *Facet) {
	self.Facets = append(self.Facets, facet)
	self.AddRoute(facet.Route)
}
