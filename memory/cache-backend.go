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

	entries map[platform.CacheKey]*platform.CacheEntry
	groups  map[string]*CacheGroup
	lock    sync.RWMutex
	pruning chan struct{}
}

func NewMemoryCacheBackend() *MemoryCacheBackend {
	self := MemoryCacheBackend{
		entries: make(map[platform.CacheKey]*platform.CacheEntry),
		groups:  make(map[string]*CacheGroup),
		pruning: make(chan struct{}),
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
func (self *MemoryCacheBackend) Load(cacheKey platform.CacheKey) (*platform.CacheEntry, bool) {
	self.lock.RLock()
	if cacheEntry, ok := self.entries[cacheKey]; ok {
		if cacheEntry.Expired() {
			self.lock.RUnlock()
			log.Debugf("cache expired: %s|%s", cacheKey, cacheEntry)
			self.lock.Lock()
			if cacheEntry.Expired() {
				delete(self.entries, cacheKey)
			}
			self.lock.Unlock()
			return nil, false
		} else {
			self.lock.RUnlock()
			return cacheEntry, true
		}
	} else {
		self.lock.RUnlock()
		return nil, false
	}
}

// CacheBackend interface
func (self *MemoryCacheBackend) Store(cacheKey platform.CacheKey, cacheEntry *platform.CacheEntry) {
	self.lock.Lock()
	defer self.lock.Unlock()

	self.entries[cacheKey] = cacheEntry

	for _, name := range cacheEntry.Groups {
		var group *CacheGroup
		var ok bool
		if group, ok = self.groups[name]; !ok {
			group = new(CacheGroup)
			self.groups[name] = group
		}
		group.Keys = append(group.Keys, cacheKey)

		// Group expiration
		for _, key := range group.Keys {
			if entry, ok := self.entries[key]; ok {
				if entry.Expiration.After(group.Expiration) {
					group.Expiration = entry.Expiration
				}
			}
		}
	}
}

// CacheBackend interface
func (self *MemoryCacheBackend) Delete(cacheKey platform.CacheKey) {
	self.lock.Lock()
	defer self.lock.Unlock()

	delete(self.entries, cacheKey)
}

// CacheBackend interface
func (self *MemoryCacheBackend) DeleteGroup(name string) {
	self.lock.Lock()
	defer self.lock.Unlock()

	if group, ok := self.groups[name]; ok {
		for _, cacheKey := range group.Keys {
			delete(self.entries, cacheKey)
		}
	}
}

func (self *MemoryCacheBackend) Prune() {
	self.lock.Lock()
	defer self.lock.Unlock()

	for cacheKey, cacheEntry := range self.entries {
		if cacheEntry.Expired() {
			log.Debugf("pruning cache: %s", cacheKey)
			delete(self.entries, cacheKey)
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
