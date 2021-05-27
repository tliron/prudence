package memory

// TODO:
// https://github.com/coocood/freecache
// https://github.com/allegro/bigcache
// https://github.com/muesli/cache2go
// https://github.com/bluele/gcache

// https://github.com/golang/groupcache

import (
	"sync"
	"time"

	"github.com/tliron/kutil/ard"
	"github.com/tliron/kutil/logging"
	"github.com/tliron/kutil/util"
	"github.com/tliron/prudence/platform"
)

var log = logging.GetLogger("prudence.memory")

func init() {
	platform.RegisterType("cache.memory", CreateMemoryCacheBackend)
}

//
// MemoryCacheBackend
//

type MemoryCacheBackend struct {
	MaxSize int // TODO

	representations map[platform.CacheKey]*platform.CachedRepresentation
	groups          map[string]*CacheGroup
	lock            sync.RWMutex
	pruning         chan struct{}
}

func NewMemoryCacheBackend() *MemoryCacheBackend {
	self := MemoryCacheBackend{
		representations: make(map[platform.CacheKey]*platform.CachedRepresentation),
		groups:          make(map[string]*CacheGroup),
		pruning:         make(chan struct{}),
	}
	self.StartPruning(10.0)
	util.OnExit(self.StopPruning)
	return &self
}

// CreateFunc signature
func CreateMemoryCacheBackend(config ard.StringMap, getRelativeURL platform.GetRelativeURL) (interface{}, error) {
	return NewMemoryCacheBackend(), nil
}

// CacheBackend interface
func (self *MemoryCacheBackend) LoadRepresentation(key platform.CacheKey) (*platform.CachedRepresentation, bool) {
	self.lock.RLock()
	if cached, ok := self.representations[key]; ok {
		if cached.Expired() {
			self.lock.RUnlock()
			log.Debugf("cache expired: %s|%s", key, cached)
			self.lock.Lock()
			if cached.Expired() {
				delete(self.representations, key)
			}
			self.lock.Unlock()
			return nil, false
		} else {
			self.lock.RUnlock()
			return cached, true
		}
	} else {
		self.lock.RUnlock()
		return nil, false
	}
}

// CacheBackend interface
func (self *MemoryCacheBackend) StoreRepresentation(cacheKey platform.CacheKey, cached *platform.CachedRepresentation) {
	self.lock.Lock()
	defer self.lock.Unlock()

	self.representations[cacheKey] = cached

	for _, name := range cached.Groups {
		var group *CacheGroup
		var ok bool
		if group, ok = self.groups[name]; !ok {
			group = new(CacheGroup)
			self.groups[name] = group
		}
		group.Keys = append(group.Keys, cacheKey)

		// Group expiration
		for _, key := range group.Keys {
			if entry, ok := self.representations[key]; ok {
				if entry.Expiration.After(group.Expiration) {
					group.Expiration = entry.Expiration
				}
			}
		}
	}
}

// CacheBackend interface
func (self *MemoryCacheBackend) DeleteRepresentation(key platform.CacheKey) {
	self.lock.Lock()
	defer self.lock.Unlock()

	delete(self.representations, key)
}

// CacheBackend interface
func (self *MemoryCacheBackend) DeleteGroup(name string) {
	self.lock.Lock()
	defer self.lock.Unlock()

	if group, ok := self.groups[name]; ok {
		for _, cacheKey := range group.Keys {
			delete(self.representations, cacheKey)
		}
	}
}

func (self *MemoryCacheBackend) Prune() {
	self.lock.Lock()
	defer self.lock.Unlock()

	for key, cached := range self.representations {
		if cached.Expired() {
			log.Debugf("pruning representation: %s", key)
			delete(self.representations, key)
		}
	}
	for name, group := range self.groups {
		if group.Expired() {
			log.Debugf("pruning group: %s", name)
			delete(self.groups, name)
		}
	}
}

func (self *MemoryCacheBackend) StartPruning(seconds float64) {
	ticker := time.NewTicker(time.Duration(seconds * 1000000000.0)) // seconds to nanoseconds
	go func() {
		for {
			select {
			case <-ticker.C:
				self.Prune()

			case <-self.pruning:
				ticker.Stop()
				return
			}
		}
	}()
}

func (self *MemoryCacheBackend) StopPruning() {
	close(self.pruning)
}

//
// CacheGroup
//

type CacheGroup struct {
	Keys       []platform.CacheKey
	Expiration time.Time
}

func (self *CacheGroup) Expired() bool {
	return time.Now().After(self.Expiration)
}
