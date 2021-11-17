package distributed

// https://github.com/buraksezer/olric

import (
	contextpkg "context"
	"sync"
	"time"

	"github.com/buraksezer/olric"
	configpkg "github.com/buraksezer/olric/config"
	"github.com/tliron/kutil/ard"
	"github.com/tliron/kutil/js"
	"github.com/tliron/kutil/logging"
	"github.com/tliron/prudence/platform"
)

var log = logging.GetLogger("prudence.distributed")

func init() {
	platform.RegisterType("DistributedCache", CreateDistributedCacheBackend)
}

//
// DistributedCacheBackend
//

type DistributedCacheBackend struct {
	Name string

	dmap *olric.DMap
	lock sync.Mutex

	olric  *olric.Olric
	config *configpkg.Config
}

func NewDistributedCacheBackend(name string) *DistributedCacheBackend {
	self := DistributedCacheBackend{Name: name}
	return &self
}

// CreateFunc signature
func CreateDistributedCacheBackend(config ard.StringMap, context *js.Context) (interface{}, error) {
	name := "prudence"
	env := "local"
	self := NewDistributedCacheBackend(name)
	if err := self.configure(env); err == nil {
		return self, nil
	} else {
		return nil, err
	}
}

// CacheBackend interface
func (self *DistributedCacheBackend) LoadRepresentation(key platform.CacheKey) (*platform.CachedRepresentation, bool) {
	self.lock.Lock()
	dmap := self.dmap
	self.lock.Unlock()

	if dmap != nil {
		if cached, err := dmap.Get(string(key)); err == nil {
			representation := cached.(platform.CachedRepresentation)
			return &representation, true
		} else {
			switch err {
			case olric.ErrKeyNotFound:
				return nil, false
			default:
				log.Errorf("%s", err)
			}
		}
	}

	return nil, false
}

// CacheBackend interface
func (self *DistributedCacheBackend) StoreRepresentation(key platform.CacheKey, cached *platform.CachedRepresentation) {
	self.lock.Lock()
	dmap := self.dmap
	self.lock.Unlock()

	if dmap != nil {
		if err := dmap.PutEx(string(key), *cached, time.Until(cached.Expiration)); err != nil {
			log.Errorf("%s", err)
		}
	}
}

// CacheBackend interface
func (self *DistributedCacheBackend) DeleteRepresentation(key platform.CacheKey) {
	self.lock.Lock()
	dmap := self.dmap
	self.lock.Unlock()

	if dmap != nil {
		if err := dmap.Delete(string(key)); err != nil {
			log.Errorf("%s", err)
		}
	}
}

// CacheBackend interface
func (self *DistributedCacheBackend) DeleteGroup(name platform.CacheKey) {
}

// Startable interface
func (self *DistributedCacheBackend) Start() error {
	return self.olric.Start()
}

// Startable interface
func (self *DistributedCacheBackend) Stop(stopContext contextpkg.Context) error {
	self.lock.Lock()
	self.dmap = nil
	olric_ := self.olric
	if olric_ != nil {
		self.olric = nil
	}
	self.lock.Unlock()

	if olric_ != nil {
		return olric_.Shutdown(stopContext)
	} else {
		return nil
	}
}

func (self *DistributedCacheBackend) configure(env string) error {
	self.config = configpkg.New(env)
	self.config.Started = self.started
	self.config.LogLevel = "DEBUG"
	//self.config.Logger =

	var err error
	if self.olric, err = olric.New(self.config); err != nil {
		return err
	}

	return nil
}

func (self *DistributedCacheBackend) started() {
	log.Notice("cache started")

	self.lock.Lock()
	defer self.lock.Unlock()

	var err error
	if self.dmap, err = self.olric.NewDMap(self.Name); err != nil {
		log.Criticalf("%s", err)
	}
}
