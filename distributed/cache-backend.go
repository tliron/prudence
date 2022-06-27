package distributed

// See: https://github.com/asim/memberlist/blob/master/memberlist.go

// https://github.com/iwanbk/bcache
// https://github.com/mailgun/groupcache
// https://github.com/iwanbk/rimcu

import (
	contextpkg "context"
	"fmt"
	"os"
	"time"

	"github.com/hashicorp/memberlist"
	"github.com/tliron/kutil/ard"
	"github.com/tliron/kutil/js"
	"github.com/tliron/kutil/kubernetes"
	"github.com/tliron/prudence/platform"
)

func init() {
	platform.RegisterType("DistributedCache", CreateDistributedCacheBackend)
}

//
// DistributedCacheBackend
//

type DistributedCacheBackend struct {
	local               platform.CacheBackend
	cluster             *memberlist.Memberlist
	queue               *memberlist.TransmitLimitedQueue
	kubernetesConfig    *KubernetesConfig
	kubernetesDiscovery *kubernetes.MemberlistPodDiscovery
}

func NewDistributedCacheBackend() *DistributedCacheBackend {
	return &DistributedCacheBackend{}
}

// platform.CreateFunc signature
func CreateDistributedCacheBackend(config ard.StringMap, context *js.Context) (interface{}, error) {
	self := NewDistributedCacheBackend()

	config_ := ard.NewNode(config)
	local := config_.Get("local").Value
	var ok bool
	if self.local, ok = local.(platform.CacheBackend); !ok {
		return nil, fmt.Errorf("DistributedCache \"local\" is not a CacheBackend: %T", local)
	}

	if kubernetes_ := config_.Get("kubernetes"); kubernetes_.Value != nil {
		self.kubernetesConfig = new(KubernetesConfig)
		self.kubernetesConfig.Namespace, _ = kubernetes_.Get("namespace").String()
		self.kubernetesConfig.Selector, _ = kubernetes_.Get("selector").String()
	}

	self.queue = &memberlist.TransmitLimitedQueue{
		NumNodes:       self.numNodes,
		RetransmitMult: 3,
	}

	config__ := memberlist.DefaultLocalConfig()
	config__.Name, _ = os.Hostname()
	config__.Delegate = self
	config__.Events = EventsDebug{}

	var err error
	if self.cluster, err = memberlist.Create(config__); err == nil {
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
	if self.kubernetesConfig != nil {
		node := self.cluster.LocalNode()
		log.Infof("starting Kubernetes discovery from %s:%d", node.Addr, node.Port)

		var err error
		self.kubernetesDiscovery, err = kubernetes.StartMemberlistPodDiscovery(self.cluster, self.kubernetesConfig.Namespace, self.kubernetesConfig.Selector, 10, log)
		return err
	}

	return nil
}

// platform.Startable interface
func (self *DistributedCacheBackend) Stop(stopContext contextpkg.Context) error {
	log.Info("stopping distributed cache")
	if self.kubernetesDiscovery != nil {
		log.Info("stopping Kubernetes discovery")
		self.kubernetesDiscovery.Stop()
	}
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

//
// KubernetesConfig
//

type KubernetesConfig struct {
	Namespace string
	Selector  string
}
