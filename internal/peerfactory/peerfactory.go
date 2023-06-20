package peerfactory

import (
	"github.com/levilutz/basiccoin/internal/pubsub"
	"github.com/levilutz/basiccoin/pkg/topic"
)

// The peer factory's subscriptions.
type subcriptions struct {
	PeerClosing *topic.SubCh[pubsub.PeerClosingEvent]
}

// A peer factory. Does not manage the peers after creation.
// May listen for inbound connections and/or seek new outbound connections.
// Keeps track of what peers exist.
type PeerFactory struct {
	pubSub *pubsub.PubSub
	subs   *subcriptions
}

// Create a new peer factory given a message bus instance.
func NewPeerFactory(pubSub *pubsub.PubSub) *PeerFactory {
	subs := &subcriptions{
		PeerClosing: pubSub.PeerClosing.SubCh(),
	}
	return &PeerFactory{
		pubSub: pubSub,
		subs:   subs,
	}
}
