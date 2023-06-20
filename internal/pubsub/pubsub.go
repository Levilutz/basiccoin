package pubsub

import "github.com/levilutz/basiccoin/pkg/topic"

// The set of pub sub topics any component needs.
type PubSub struct {
	PeerAnnouncedAddr  *topic.Topic[PeerAnnouncedAddrEvent]
	PeerClosing        *topic.Topic[PeerClosingEvent]
	PeersReceived      *topic.Topic[PeersReceivedEvent]
	PeersRequested     *topic.Topic[PeersRequestedEvent]
	PrintUpdate        *topic.Topic[PrintUpdateEvent]
	SendPeers          *topic.Topic[SendPeersEvent]
	ShouldAnnounceAddr *topic.Topic[ShouldAnnounceAddrEvent]
	ShouldRequestPeers *topic.Topic[ShouldRequestPeersEvent]
	ValidatedHead      *topic.Topic[ValidatedHeadEvent]
}

func NewPubSub() *PubSub {
	return &PubSub{
		PeerAnnouncedAddr:  topic.NewTopic[PeerAnnouncedAddrEvent](),
		PeerClosing:        topic.NewTopic[PeerClosingEvent](),
		PeersReceived:      topic.NewTopic[PeersReceivedEvent](),
		PeersRequested:     topic.NewTopic[PeersRequestedEvent](),
		PrintUpdate:        topic.NewTopic[PrintUpdateEvent](),
		SendPeers:          topic.NewTopic[SendPeersEvent](),
		ShouldAnnounceAddr: topic.NewTopic[ShouldAnnounceAddrEvent](),
		ShouldRequestPeers: topic.NewTopic[ShouldRequestPeersEvent](),
		ValidatedHead:      topic.NewTopic[ValidatedHeadEvent](),
	}
}
