package pubsub

import "github.com/levilutz/basiccoin/pkg/topic"

// The set of pub sub topics any component needs.
type PubSub struct {
	PeerClosing        *topic.Topic[PeerClosingEvent]
	PeersRequested     *topic.Topic[PeersRequestedEvent]
	SendPeers          *topic.Topic[SendPeersEvent]
	ShouldRequestPeers *topic.Topic[ShouldRequestPeersEvent]
	ValidatedHead      *topic.Topic[ValidatedHeadEvent]
}

func NewPubSub() *PubSub {
	return &PubSub{
		PeerClosing:        topic.NewTopic[PeerClosingEvent](),
		PeersRequested:     topic.NewTopic[PeersRequestedEvent](),
		SendPeers:          topic.NewTopic[SendPeersEvent](),
		ShouldRequestPeers: topic.NewTopic[ShouldRequestPeersEvent](),
		ValidatedHead:      topic.NewTopic[ValidatedHeadEvent](),
	}
}
