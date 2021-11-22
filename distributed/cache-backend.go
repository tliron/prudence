package distributed

// https://github.com/iwanbk/bcache
// https://github.com/hashicorp/memberlist
// https://github.com/mailgun/groupcache
// https://github.com/iwanbk/rimcu

// See: https://github.com/asim/memberlist/blob/master/memberlist.go

import (
	contextpkg "context"
	"os"
	"time"

	"github.com/hashicorp/memberlist"
	"github.com/tliron/kutil/ard"
	"github.com/tliron/kutil/js"
	"github.com/tliron/prudence/platform"
)

func init() {
	platform.RegisterType("DistributedCache", CreateDistributedCacheBackend)
}

//
// DistributedCacheBackend
//

type DistributedCacheBackend struct {
	local   platform.CacheBackend
	cluster *memberlist.Memberlist
	queue   *memberlist.TransmitLimitedQueue
}

func NewDistributedCacheBackend() *DistributedCacheBackend {
	return &DistributedCacheBackend{}
}

// platform.CreateFunc signature
func CreateDistributedCacheBackend(config ard.StringMap, context *js.Context) (interface{}, error) {
	self := NewDistributedCacheBackend()

	if b, err := platform.Create(ard.StringMap{"type": "MemoryCache"}, context); err == nil {
		self.local = b.(platform.CacheBackend)
	} else {
		return nil, err
	}

	self.queue = &memberlist.TransmitLimitedQueue{
		NumNodes:       self.numNodes,
		RetransmitMult: 3,
	}

	config_ := memberlist.DefaultLocalConfig()
	config_.Name, _ = os.Hostname()
	config_.Delegate = self
	config_.Events = EventsDebug{}

	var err error
	if self.cluster, err = memberlist.Create(config_); err == nil {
		return self, nil
	} else {
		return nil, err
	}
}

// platform.CacheBackend interface
func (self *DistributedCacheBackend) LoadRepresentation(key platform.CacheKey) (*platform.CachedRepresentation, bool) {
	return self.local.LoadRepresentation(key)
}

// platform.CacheBackend interface
func (self *DistributedCacheBackend) StoreRepresentation(key platform.CacheKey, cached *platform.CachedRepresentation) {
	self.local.StoreRepresentation(key, cached)
	self.queue.QueueBroadcast(NewStoreRepresentationMessage(key, cached))
}

// platform.CacheBackend interface
func (self *DistributedCacheBackend) DeleteRepresentation(key platform.CacheKey) {
	self.local.DeleteRepresentation(key)
	self.queue.QueueBroadcast(NewDeleteRepresentationMessage(key))
}

// platform.CacheBackend interface
func (self *DistributedCacheBackend) DeleteGroup(name platform.CacheKey) {
	self.local.DeleteGroup(name)
	self.queue.QueueBroadcast(NewDeleteGroupMessage(name))
}

// platform.Startable interface
func (self *DistributedCacheBackend) Start() error {
	node := self.cluster.LocalNode()
	nodes := DiscoverKubernetesNodes()
	log.Infof("starting distributed cache on %s:%d: %v", node.Addr, node.Port, nodes)
	_, err := self.cluster.Join(nodes)
	return err
}

// platform.Startable interface
func (self *DistributedCacheBackend) Stop(stopContext contextpkg.Context) error {
	log.Info("stopping distributed cache")
	err := self.cluster.Leave(time.Second * 5)
	self.cluster.Shutdown()
	return err
}

// memberlist.Delegate interface
func (self *DistributedCacheBackend) NodeMeta(limit int) []byte {
	return nil
}

// memberlist.Delegate interface
func (self *DistributedCacheBackend) NotifyMsg(bytes []byte) {
	if message := ParseMessage(bytes); message != nil {
		switch message.Type {
		case StoreRepresentationMessageType:
			log.Debugf("remote store: %s", message.Key)
			self.local.StoreRepresentation(message.Key, message.Representation)
		case DeleteRepresentationMessageType:
			log.Debugf("remote delete: %s", message.Key)
			self.local.DeleteRepresentation(message.Key)
		case DeleteGroupMessageType:
			log.Debugf("remote delete group: %s", message.Key)
			self.local.DeleteGroup(message.Key)
		}
	}
}

// memberlist.Delegate interface
func (self *DistributedCacheBackend) GetBroadcasts(overhead int, limit int) [][]byte {
	return self.queue.GetBroadcasts(overhead, limit)
}

// memberlist.Delegate interface
func (self *DistributedCacheBackend) LocalState(join bool) []byte {
	return nil
}

// memberlist.Delegate interface
func (self *DistributedCacheBackend) MergeRemoteState(buf []byte, join bool) {
}

func (self *DistributedCacheBackend) numNodes() int {
	return self.cluster.NumMembers()
}
