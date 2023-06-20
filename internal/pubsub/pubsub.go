package pubsub

import "github.com/levilutz/basiccoin/pkg/topic"

// The set of pub sub topics any component needs.
type PubSub struct {
	PeerClosing        *topic.Topic[PeerClosingEvent]
	ShouldRequestPeers *topic.Topic[ShouldRequestPeersEvent]
	ValidatedHead      *topic.Topic[ValidatedHeadEvent]
}

func NewPubSub() *PubSub {
	return &PubSub{
		PeerClosing:        topic.NewTopic[PeerClosingEvent](),
		ValidatedHead:      topic.NewTopic[ValidatedHeadEvent](),
		ShouldRequestPeers: topic.NewTopic[ShouldRequestPeersEvent](),
	}
}
