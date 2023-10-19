package memory

import (
	"sync"
	"time"

	"github.com/tliron/commonjs-goja"
	"github.com/tliron/go-ard"
	"github.com/tliron/kutil/util"
	"github.com/tliron/prudence/platform"
)

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

// ([platform.CreateFunc] signature)
func CreateMapCacheBackend(jsContext *commonjs.Context, config ard.StringMap) (any, error) {
	self := NewMapCacheBackend()

	var pruneFrequency float64

	config_ := ard.With(config).ConvertSimilar().NilMeansZero()
	var ok bool
	if pruneFrequency, ok = config_.Get("pruneFrequency").Float(); !ok {
		pruneFrequency = 10.0 // seconds
	}

	self.StartPruning(pruneFrequency)
	util.OnExit(self.StopPruning)
	return self, nil
}

// ([platform.CacheBackend] interface)
func (self *MapCacheBackend) LoadRepresentation(key platform.CacheKey) (*platform.CachedRepresentation, bool) {
	self.lock.RLock()
	if cached, ok := self.representations[key]; ok {
		if cached.Expired() {
			self.lock.RUnlock()
			log.Debug("cache expired", "key", key, "encodings", cached.String())
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

// ([platform.CacheBackend] interface)
func (self *MapCacheBackend) StoreRepresentation(key platform.CacheKey, cached *platform.CachedRepresentation) {
	go func() {
		self.lock.Lock()
		defer self.lock.Unlock()

		self.representations[key] = cached
		self.groups.Add(key, cached, self.getExpiration)
	}()
}

// ([platform.CacheBackend] interface)
func (self *MapCacheBackend) DeleteRepresentation(key platform.CacheKey) {
	go func() {
		self.lock.Lock()
		defer self.lock.Unlock()

		delete(self.representations, key)
	}()
}

// ([platform.CacheBackend] interface)
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
			log.Debug("pruning representation", "key", key)
			delete(self.representations, key)
		}
	}

	self.groups.Prune(self.getExpiration)
}

func (self *MapCacheBackend) StartPruning(frequencySeconds float64) {
	ticker := time.NewTicker(time.Duration(frequencySeconds * float64(time.Second)))
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
