package peerfactory

import (
	"fmt"
	"time"

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
	params Params
	pubSub *pubsub.PubSub
	subs   *subcriptions
}

// Create a new peer factory given a message bus instance.
func NewPeerFactory(params Params, pubSub *pubsub.PubSub) *PeerFactory {
	subs := &subcriptions{
		PeerClosing: pubSub.PeerClosing.SubCh(),
	}
	return &PeerFactory{
		params: params,
		pubSub: pubSub,
		subs:   subs,
	}
}

// Start the peer factory's loop.
func (pf *PeerFactory) Loop() {
	seekPeersTicker := time.NewTicker(pf.params.SeekNewPeersFreq)
	for {
		select {
		case peerClosingEvent := <-pf.subs.PeerClosing.C:
			fmt.Println("peer closing received:", peerClosingEvent.PeerRuntimeId)
		case <-seekPeersTicker.C:
			fmt.Println("check if we need new peers")
		}
	}
}
