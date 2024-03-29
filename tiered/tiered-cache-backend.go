package tiered

import (
	"fmt"

	"github.com/tliron/commonjs-goja"
	"github.com/tliron/go-ard"
	"github.com/tliron/prudence/platform"
)

//
// TieredCacheBackend
//

type TieredCacheBackend struct {
	cacheBackends []platform.CacheBackend
}

func NewTieredCacheBackend() *TieredCacheBackend {
	return &TieredCacheBackend{}
}

// ([platform.CreateFunc] signature)
func CreateTieredCacheBackend(jsContext *commonjs.Context, config ard.StringMap) (any, error) {
	self := NewTieredCacheBackend()

	config_ := ard.With(config).ConvertSimilar().NilMeansZero()
	if list, ok := config_.Get("caches").List(); ok {
		for _, cache := range list {
			if cacheBackend, ok := cache.(platform.CacheBackend); ok {
				self.cacheBackends = append(self.cacheBackends, cacheBackend)
			} else {
				return nil, fmt.Errorf("TieredCache \"caches\" contains an object that is not a cache backend: %T", cache)
			}
		}
	}

	return self, nil
}

// ([platform.CacheBackend] interface)
func (self *TieredCacheBackend) LoadRepresentation(key platform.CacheKey) (*platform.CachedRepresentation, bool) {
	for index, cacheBackend := range self.cacheBackends {
		if cached, ok := cacheBackend.LoadRepresentation(key); ok {
			// Store in previous tiers
			for i := 0; i < index; i++ {
				self.cacheBackends[i].StoreRepresentation(key, cached)
			}

			return cached, true
		}
	}
	return nil, false
}

// ([platform.CacheBackend] interface)
func (self *TieredCacheBackend) StoreRepresentation(key platform.CacheKey, cached *platform.CachedRepresentation) {
	for _, cacheBackend := range self.cacheBackends {
		cacheBackend.StoreRepresentation(key, cached)
	}
}

// ([platform.CacheBackend] interface)
func (self *TieredCacheBackend) DeleteRepresentation(key platform.CacheKey) {
	for _, cacheBackend := range self.cacheBackends {
		cacheBackend.DeleteRepresentation(key)
	}
}

// ([platform.CacheBackend] interface)
func (self *TieredCacheBackend) DeleteGroup(name platform.CacheKey) {
	for _, cacheBackend := range self.cacheBackends {
		cacheBackend.DeleteGroup(name)
	}
}

// platform.HasStartables interface
func (self *TieredCacheBackend) GetStartables() []platform.Startable {
	var startables []platform.Startable
	for _, cacheBackend := range self.cacheBackends {
		if startable, ok := cacheBackend.(platform.Startable); ok {
			startables = append(startables, startable)
		}
	}
	return startables
}
