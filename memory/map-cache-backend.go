package memory

import (
	"sync"
	"time"

	"github.com/tliron/kutil/ard"
	"github.com/tliron/kutil/js"
	"github.com/tliron/kutil/util"
	"github.com/tliron/prudence/platform"
)

func init() {
	platform.RegisterType("MapCache", CreateMapCacheBackend)
}

//
// MapCacheBackend
//

type MapCacheBackend struct {
	representations map[platform.CacheKey]*platform.CachedRepresentation
	groups          CacheGroups
	lock            sync.RWMutex
	pruning         chan struct{}
}

func NewMapCacheBackend() *MapCacheBackend {
	return &MapCacheBackend{
		representations: make(map[platform.CacheKey]*platform.CachedRepresentation),
		groups:          make(CacheGroups),
		pruning:         make(chan struct{}),
	}
}

// platform.CreateFunc signature
func CreateMapCacheBackend(config ard.StringMap, context *js.Context) (interface{}, error) {
	self := NewMapCacheBackend()

	var pruneFrequency float64

	config_ := ard.NewNode(config)
	var ok bool
	if pruneFrequency, ok = config_.Get("pruneFrequency").Float(); !ok {
		pruneFrequency = 10.0 // seconds
	}

	self.StartPruning(pruneFrequency)
	util.OnExit(self.StopPruning)
	return self, nil
}

// platform.CacheBackend interface
func (self *MapCacheBackend) LoadRepresentation(key platform.CacheKey) (*platform.CachedRepresentation, bool) {
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

// platform.CacheBackend interface
func (self *MapCacheBackend) StoreRepresentation(key platform.CacheKey, cached *platform.CachedRepresentation) {
	go func() {
		self.lock.Lock()
		defer self.lock.Unlock()

		self.representations[key] = cached
		self.groups.Add(key, cached, self.getExpiration)
	}()
}

// platform.CacheBackend interface
func (self *MapCacheBackend) DeleteRepresentation(key platform.CacheKey) {
	go func() {
		self.lock.Lock()
		defer self.lock.Unlock()

		delete(self.representations, key)
	}()
}

// platform.CacheBackend interface
func (self *MapCacheBackend) DeleteGroup(name platform.CacheKey) {
	go func() {
		self.lock.Lock()
		defer self.lock.Unlock()
		self.groups.Delete(name, func(key platform.CacheKey) {
			delete(self.representations, key)
		})
	}()
}

func (self *MapCacheBackend) Prune() {
	self.lock.Lock()
	defer self.lock.Unlock()

	for key, cached := range self.representations {
		if cached.Expired() {
			log.Debugf("pruning representation: %s", key)
			delete(self.representations, key)
		}
	}

	self.groups.Prune(self.getExpiration)
}

func (self *MapCacheBackend) StartPruning(frequency float64) {
	ticker := time.NewTicker(time.Duration(frequency * 1000000000.0)) // seconds to nanoseconds
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

func (self *MapCacheBackend) StopPruning() {
	close(self.pruning)
}

// GetExpirationFunc signature
func (self *MapCacheBackend) getExpiration(key platform.CacheKey) (time.Time, bool) {
	if cached, ok := self.representations[key]; ok {
		if !cached.Expired() {
			return cached.Expiration, true
		} else {
			return time.Time{}, false
		}
	} else {
		return time.Time{}, false
	}
}
