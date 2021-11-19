package tiered

import (
	"fmt"

	"github.com/tliron/kutil/ard"
	"github.com/tliron/kutil/js"
	"github.com/tliron/kutil/logging"
	"github.com/tliron/prudence/platform"
)

var log = logging.GetLogger("prudence.tiered")

func init() {
	platform.RegisterType("TieredCache", CreateTieredCacheBackend)
}

//
// TieredCacheBackend
//

type TieredCacheBackend struct {
	cacheBackends []platform.CacheBackend
}

func NewTieredCacheBackend() *TieredCacheBackend {
	return &TieredCacheBackend{}
}

// platform.CreateFunc signature
func CreateTieredCacheBackend(config ard.StringMap, context *js.Context) (interface{}, error) {
	self := NewTieredCacheBackend()
	config_ := ard.NewNode(config)
	if list, ok := config_.Get("caches").List(false); ok {
		for _, cache := range list {
			if cacheBackend, ok := cache.(platform.CacheBackend); ok {
				self.cacheBackends = append(self.cacheBackends, cacheBackend)
			} else {
				return nil, fmt.Errorf("not a cache backend: %T", cache)
			}
		}
	}
	return self, nil
}

// platform.CacheBackend interface
func (self *TieredCacheBackend) LoadRepresentation(key platform.CacheKey) (*platform.CachedRepresentation, bool) {
	for _, cacheBackend := range self.cacheBackends {
		if representation, ok := cacheBackend.LoadRepresentation(key); ok {
			return representation, true
		}
	}
	return nil, false
}

// platform.CacheBackend interface
func (self *TieredCacheBackend) StoreRepresentation(key platform.CacheKey, cached *platform.CachedRepresentation) {
	for _, cacheBackend := range self.cacheBackends {
		cacheBackend.StoreRepresentation(key, cached)
	}
}

// platform.CacheBackend interface
func (self *TieredCacheBackend) DeleteRepresentation(key platform.CacheKey) {
	for _, cacheBackend := range self.cacheBackends {
		cacheBackend.DeleteRepresentation(key)
	}
}

// platform.CacheBackend interface
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
