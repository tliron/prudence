package distributed

import (
	"github.com/fxamacker/cbor/v2"
	"github.com/hashicorp/memberlist"
	"github.com/tliron/prudence/platform"
)

type MessageType int

const (
	StoreRepresentationMessageType  = MessageType(1)
	DeleteRepresentationMessageType = MessageType(2)
	DeleteGroupMessageType          = MessageType(3)
)

//
// Message
//

type Message struct {
	Type           MessageType
	Key            platform.CacheKey
	Representation *platform.CachedRepresentation
}

func NewStoreRepresentationMessage(key platform.CacheKey, cached *platform.CachedRepresentation) *Message {
	return &Message{
		Type:           StoreRepresentationMessageType,
		Key:            key,
		Representation: cached,
	}
}

func NewDeleteRepresentationMessage(key platform.CacheKey) *Message {
	return &Message{
		Type: DeleteRepresentationMessageType,
		Key:  key,
	}
}

func NewDeleteGroupMessage(name platform.CacheKey) *Message {
	return &Message{
		Type: DeleteGroupMessageType,
		Key:  name,
	}
}

func ParseMessage(bytes []byte) *Message {
	var self Message
	if err := cbor.Unmarshal(bytes, &self); err == nil {
		return &self
	} else {
		return nil
	}
}

// ([memberlist.Broadcast] interface)
func (self *Message) Invalidates(broadcast memberlist.Broadcast) bool {
	return false
}

// ([memberlist.Broadcast] interface)
func (self *Message) Message() []byte {
	bytes, _ := cbor.Marshal(self)
	return bytes
}

// ([memberlist.Broadcast] interface)
func (self *Message) Finished() {
}
