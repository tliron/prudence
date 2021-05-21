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

	cache   map[platform.CacheKey]*platform.CacheEntry
	lock    sync.RWMutex
	pruning chan struct{}
}

func NewMemoryCacheBackend() *MemoryCacheBackend {
	self := MemoryCacheBackend{
		cache:   make(map[platform.CacheKey]*platform.CacheEntry),
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
	if cacheEntry, ok := self.cache[cacheKey]; ok {
		if cacheEntry.Expired() {
			self.lock.RUnlock()
			log.Debugf("cache expired: %s|%s", cacheKey, cacheEntry)
			self.lock.Lock()
			if cacheEntry.Expired() {
				delete(self.cache, cacheKey)
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
	self.cache[cacheKey] = cacheEntry
	self.lock.Unlock()
}

// CacheBackend interface
func (self *MemoryCacheBackend) Delete(cacheKey platform.CacheKey) {
	self.lock.Lock()
	delete(self.cache, cacheKey)
	self.lock.Unlock()
}

func (self *MemoryCacheBackend) Prune() {
	self.lock.Lock()
	for cacheKey, cacheEntry := range self.cache {
		if cacheEntry.Expired() {
			log.Debugf("pruning cache: %s", cacheKey)
			delete(self.cache, cacheKey)
		}
	}
	self.lock.Unlock()
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
