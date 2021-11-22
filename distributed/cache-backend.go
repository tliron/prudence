package distributed

// https://github.com/iwanbk/bcache
// https://github.com/mailgun/groupcache
// https://github.com/iwanbk/rimcu

import (
	contextpkg "context"

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
}

func NewDistributedCacheBackend() *DistributedCacheBackend {
	return &DistributedCacheBackend{}
}

// platform.CreateFunc signature
func CreateDistributedCacheBackend(config ard.StringMap, context *js.Context) (interface{}, error) {
	self := NewDistributedCacheBackend()
	return self, nil
}

// platform.CacheBackend interface
func (self *DistributedCacheBackend) LoadRepresentation(key platform.CacheKey) (*platform.CachedRepresentation, bool) {
	return nil, false
}

// platform.CacheBackend interface
func (self *DistributedCacheBackend) StoreRepresentation(key platform.CacheKey, cached *platform.CachedRepresentation) {
}

// platform.CacheBackend interface
func (self *DistributedCacheBackend) DeleteRepresentation(key platform.CacheKey) {
}

// platform.CacheBackend interface
func (self *DistributedCacheBackend) DeleteGroup(name platform.CacheKey) {
}

// platform.Startable interface
func (self *DistributedCacheBackend) Start() error {
	log.Info("starting distributed cache")
	return nil
}

// platform.Startable interface
func (self *DistributedCacheBackend) Stop(stopContext contextpkg.Context) error {
	log.Info("stopping distributed cache")
	return nil
}
