package memory

// See: https://github.com/gostor/awesome-go-storage
// TODO: https://github.com/kelindar/column

import (
	"sync"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/tliron/commonjs-goja"
	"github.com/tliron/go-ard"
	"github.com/tliron/kutil/util"
	"github.com/tliron/prudence/platform"
)

//
// MemoryCacheBackend
//

type MemoryCacheBackend struct {
	cache   *ristretto.Cache
	groups  CacheGroups
	lock    sync.RWMutex
	pruning chan struct{}
}

func NewMemoryCacheBackend() *MemoryCacheBackend {
	return &MemoryCacheBackend{
		groups:  make(CacheGroups),
		pruning: make(chan struct{}),
	}
}

// ([platform.CreateFunc] signature)
func CreateMemoryCacheBackend(jsContext *commonjs.Context, config ard.StringMap) (any, error) {
	self := NewMemoryCacheBackend()

	var maxSize int64
	var averageSize int64
	var pruneFrequency float64

	config_ := ard.With(config).ConvertSimilar().NilMeansZero()
	var ok bool
	if maxSize, ok = config_.Get("maxSize").Integer(); !ok {
		maxSize = 1073741824 // 1 GiB
	}
	if averageSize, ok = config_.Get("averageSize").Integer(); !ok {
		averageSize = 1000
	}
	if pruneFrequency, ok = config_.Get("pruneFrequency").Float(); !ok {
		pruneFrequency = 10.0 // seconds
	}

	config__ := ristretto.Config{
		MaxCost: maxSize,
		// Recommendations:
		BufferItems: 64,
		NumCounters: 100 * (maxSize / averageSize),
	}

	var err error
	if self.cache, err = ristretto.NewCache(&config__); err == nil {
		util.OnExit(self.cache.Close)
		self.StartPruning(pruneFrequency)
		util.OnExit(self.StopPruning)
		return self, nil
	} else {
		return nil, err
	}
}

// ([platform.CacheBackend] interface)
func (self *MemoryCacheBackend) LoadRepresentation(key platform.CacheKey) (*platform.CachedRepresentation, bool) {
	if cached, ok := self.cache.Get(string(key)); ok {
		return cached.(*platform.CachedRepresentation), true
	} else {
		return nil, false
	}
}

// ([platform.CacheBackend] interface)
func (self *MemoryCacheBackend) StoreRepresentation(key platform.CacheKey, cached *platform.CachedRepresentation) {
	self.cache.SetWithTTL(string(key), cached, int64(cached.GetSize()), cached.Expiration.Sub(time.Now()))

	if len(cached.Groups) > 0 {
		go func() {
			self.lock.Lock()
			defer self.lock.Unlock()
			self.groups.Add(key, cached, self.getExpiration)
		}()
	}
}

// ([platform.CacheBackend] interface)
func (self *MemoryCacheBackend) DeleteRepresentation(key platform.CacheKey) {
	self.cache.Del(string(key))
}

// ([platform.CacheBackend] interface)
func (self *MemoryCacheBackend) DeleteGroup(name platform.CacheKey) {
	go func() {
		self.lock.Lock()
		defer self.lock.Unlock()
		self.groups.Delete(name, func(key platform.CacheKey) {
			self.cache.Del(string(key))
		})
	}()
}

func (self *MemoryCacheBackend) Prune() {
	self.lock.Lock()
	defer self.lock.Unlock()

	self.groups.Prune(self.getExpiration)
}

func (self *MemoryCacheBackend) StartPruning(frequencySeconds float64) {
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

func (self *MemoryCacheBackend) StopPruning() {
	close(self.pruning)
}

// GetExpirationFunc signature
func (self *MemoryCacheBackend) getExpiration(key platform.CacheKey) (time.Time, bool) {
	if cached, ok := self.cache.Get(string(key)); ok {
		cached_ := cached.(*platform.CachedRepresentation)
		if !cached_.Expired() {
			return cached_.Expiration, true
		} else {
			return time.Time{}, false
		}
	} else {
		return time.Time{}, false
	}
}
