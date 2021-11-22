package distributed

import (
	"github.com/hashicorp/memberlist"
)

//
// EventsDebug
//

type EventsDebug struct{}

// memberlist.EventDelegate interface
func (self EventsDebug) NotifyJoin(node *memberlist.Node) {
	log.Debugf("node has joined: %s", node.String())
}

// memberlist.EventDelegate interface
func (self EventsDebug) NotifyLeave(node *memberlist.Node) {
	log.Debugf("node has left: %s", node.String())
}

// memberlist.EventDelegate interface
func (self EventsDebug) NotifyUpdate(node *memberlist.Node) {
	log.Debugf("node was updated: %s", node.String())
}
