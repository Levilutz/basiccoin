package peer

import (
	"github.com/levilutz/basiccoin/internal/pubsub"
	"github.com/levilutz/basiccoin/pkg/topic"
)

// The peer's subscriptions.
type subscriptions struct {
	ValidatedHead *topic.SubCh[pubsub.ValidatedHeadEvent]
}

// Close our subscriptions as we close.
func (s subscriptions) Close() {
	s.ValidatedHead.Close()
}

// A connection to a single peer.
type Peer struct {
	pubSub *pubsub.PubSub
	subs   *subscriptions
}

// Create a new peer given a message bus instance.
func NewPeer(pubSub *pubsub.PubSub) *Peer {
	return &Peer{
		pubSub: pubSub,
		subs: &subscriptions{
			ValidatedHead: pubSub.ValidatedHead.SubCh(),
		},
	}
}
