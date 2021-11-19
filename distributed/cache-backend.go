package distributed

// https://github.com/buraksezer/olric

import (
	contextpkg "context"
	"fmt"
	logpkg "log"
	"sync"
	"time"

	"github.com/buraksezer/olric"
	configpkg "github.com/buraksezer/olric/config"
	"github.com/tliron/kutil/ard"
	"github.com/tliron/kutil/js"
	"github.com/tliron/kutil/logging"
	urlpkg "github.com/tliron/kutil/url"
	"github.com/tliron/kutil/util"
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
	representationsMapName string
	groupsMapName          string

	olric              *olric.Olric
	representationsMap *olric.DMap
	groupsMap          *olric.DMap
	lock               sync.Mutex
}

func NewDistributedCacheBackend() *DistributedCacheBackend {
	return &DistributedCacheBackend{}
}

// platform.CreateFunc signature
func CreateDistributedCacheBackend(config ard.StringMap, context *js.Context) (interface{}, error) {
	self := NewDistributedCacheBackend()

	config_ := ard.NewNode(config)
	self.representationsMapName, _ = config_.Get("representationsMap").String(true)
	if self.representationsMapName == "" {
		self.representationsMapName = "prudence.representations"
	}
	self.groupsMapName, _ = config_.Get("groupsMap").String(true)
	if self.groupsMapName == "" {
		self.groupsMapName = "prudence.groups"
	}

	var config__ *configpkg.Config
	if load, ok := config_.Get("load").String(false); ok {
		if url, err := context.Resolve(load, true); err == nil {
			if fileUrl, ok := url.(*urlpkg.FileURL); ok {
				if config__, err = configpkg.Load(fileUrl.Path); err != nil {
					return nil, fmt.Errorf("could not configure Olric: %s", err)
				}
			} else {
				return nil, fmt.Errorf("not a file: %v", url)
			}
		} else {
			return nil, err
		}
	} else {
		config__ = configpkg.New("local")
	}

	config__.Started = self.started
	config__.Logger = logpkg.New(logWriter{}, "", 0)

	log.Info("creating Olric cache")
	var err error
	if self.olric, err = olric.New(config__); err == nil {
		return self, nil
	} else {
		return nil, err
	}
}

// platform.CacheBackend interface
func (self *DistributedCacheBackend) LoadRepresentation(key platform.CacheKey) (*platform.CachedRepresentation, bool) {
	self.lock.Lock()
	representationsMap := self.representationsMap
	self.lock.Unlock()

	if representationsMap != nil {
		if cached, err := representationsMap.Get(string(key)); err == nil {
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

// platform.CacheBackend interface
func (self *DistributedCacheBackend) StoreRepresentation(key platform.CacheKey, cached *platform.CachedRepresentation) {
	self.lock.Lock()
	representationsMap := self.representationsMap
	self.lock.Unlock()

	if representationsMap != nil {
		if err := representationsMap.PutEx(string(key), *cached, time.Until(cached.Expiration)); err != nil {
			log.Errorf("%s", err)
		}
	}
}

// platform.CacheBackend interface
func (self *DistributedCacheBackend) DeleteRepresentation(key platform.CacheKey) {
	self.lock.Lock()
	representationsMap := self.representationsMap
	self.lock.Unlock()

	if representationsMap != nil {
		if err := representationsMap.Delete(string(key)); err != nil {
			log.Errorf("%s", err)
		}
	}
}

// platform.CacheBackend interface
func (self *DistributedCacheBackend) DeleteGroup(name platform.CacheKey) {
	self.lock.Lock()
	groupsMap := self.groupsMap
	self.lock.Unlock()

	if groupsMap != nil {
		// TODO
	}
}

// platform.Startable interface
func (self *DistributedCacheBackend) Start() error {
	log.Info("starting Olric cache")
	return self.olric.Start()
}

// platform.Startable interface
func (self *DistributedCacheBackend) Stop(stopContext contextpkg.Context) error {
	log.Info("stopping Olric cache")
	self.lock.Lock()
	self.representationsMap = nil
	self.groupsMap = nil
	olric_ := self.olric
	if olric_ != nil {
		self.olric = nil
	}
	self.lock.Unlock()

	if olric_ != nil {
		if err := olric_.Shutdown(stopContext); err == nil {
			log.Info("Olric cache stopped")
			return nil
		} else {
			return err
		}
	} else {
		return nil
	}
}

func (self *DistributedCacheBackend) started() {
	log.Info("Olric cache started")

	self.lock.Lock()
	defer self.lock.Unlock()

	var err error
	if self.representationsMap, err = self.olric.NewDMap(self.representationsMapName); err != nil {
		log.Criticalf("%s", err)
	}

	if self.groupsMap, err = self.olric.NewDMap(self.groupsMapName); err != nil {
		log.Criticalf("%s", err)
	}
}

//
// logWriter
//

type logWriter struct{}

func (self logWriter) Write(p []byte) (int, error) {
	// TODO: very hacky solution!
	// See: https://github.com/buraksezer/olric/issues/117#issuecomment-899289898
	message := util.BytesToString(p)
	length := len(message)
	if length < 2 {
		log.Info(message)
	} else {
		switch message[1] {
		case 'I':
			log.Info(message[7:])
		case 'W':
			log.Warning(message[7:])
		case 'D':
			log.Debug(message[8:])
		default:
			log.Info(message)
		}
	}
	return length, nil
}
