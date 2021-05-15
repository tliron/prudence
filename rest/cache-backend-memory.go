package rest

import (
	"sync"
	"time"

	"github.com/tliron/kutil/logging"
	"github.com/tliron/kutil/util"
)

var logMemoryBackend = logging.GetLogger("prudence.cache.memory")

func init() {
	cacheBackend = NewCacheBackendMemory()
}

//
// CacheBackendMemory
//

type CacheBackendMemory struct {
	cache   sync.Map
	pruning chan struct{}
}

func NewCacheBackendMemory() *CacheBackendMemory {
	self := CacheBackendMemory{
		pruning: make(chan struct{}),
	}
	self.StartPruning(10.0)
	util.OnExit(self.StopPruning)
	return &self
}

// CacheBackend interface
func (self *CacheBackendMemory) Load(cacheKey CacheKey) (*CacheEntry, bool) {
	if cacheEntry, ok := self.cache.Load(cacheKey); ok {
		cacheEntry_ := cacheEntry.(*CacheEntry)
		if cacheEntry_.Expired() {
			logMemoryBackend.Debugf("cache entry expired: %s", cacheKey)
			self.cache.Delete(cacheKey)
			return nil, false
		} else {
			logMemoryBackend.Debugf("cache hit: %s", cacheKey)
			return cacheEntry_, true
		}
	} else {
		logMemoryBackend.Debugf("not cached: %s", cacheKey)
		return nil, false
	}
}

// CacheBackend interface
func (self *CacheBackendMemory) Store(cacheKey CacheKey, cacheEntry *CacheEntry) {
	self.cache.Store(cacheKey, cacheEntry)
}

// CacheBackend interface
func (self *CacheBackendMemory) Delete(cacheKey CacheKey) {
	self.cache.Delete(cacheKey)
}

func (self *CacheBackendMemory) Prune() {
	self.cache.Range(func(key interface{}, value interface{}) bool {
		if value.(*CacheEntry).Expired() {
			logMemoryBackend.Debugf("pruning cache entry: %s", key.(CacheKey))
			self.cache.Delete(key)
		}
		return true
	})
}

func (self *CacheBackendMemory) StartPruning(seconds float64) {
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

func (self *CacheBackendMemory) StopPruning() {
	close(self.pruning)
}
