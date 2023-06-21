package pubsub

import "github.com/levilutz/basiccoin/pkg/topic"

// The set of pub sub topics any component needs.
type PubSub struct {
	CandidateHead      *topic.Topic[CandidateHeadEvent]
	CandidateTx        *topic.Topic[CandidateTxEvent]
	MinerTarget        *topic.Topic[MinerTargetEvent]
	PeerAnnouncedAddr  *topic.Topic[PeerAnnouncedAddrEvent]
	PeerClosing        *topic.Topic[PeerClosingEvent]
	PeersReceived      *topic.Topic[PeersReceivedEvent]
	PeersRequested     *topic.Topic[PeersRequestedEvent]
	PrintUpdate        *topic.Topic[PrintUpdateEvent]
	SendPeers          *topic.Topic[SendPeersEvent]
	ShouldAnnounceAddr *topic.Topic[ShouldAnnounceAddrEvent]
	ShouldRequestPeers *topic.Topic[ShouldRequestPeersEvent]
	ValidatedHead      *topic.Topic[ValidatedHeadEvent]
	ValidatedTx        *topic.Topic[ValidatedTxEvent]
}

func NewPubSub() *PubSub {
	return &PubSub{
		CandidateHead:      topic.NewTopic[CandidateHeadEvent](),
		CandidateTx:        topic.NewTopic[CandidateTxEvent](),
		MinerTarget:        topic.NewTopic[MinerTargetEvent](),
		PeerAnnouncedAddr:  topic.NewTopic[PeerAnnouncedAddrEvent](),
		PeerClosing:        topic.NewTopic[PeerClosingEvent](),
		PeersReceived:      topic.NewTopic[PeersReceivedEvent](),
		PeersRequested:     topic.NewTopic[PeersRequestedEvent](),
		PrintUpdate:        topic.NewTopic[PrintUpdateEvent](),
		SendPeers:          topic.NewTopic[SendPeersEvent](),
		ShouldAnnounceAddr: topic.NewTopic[ShouldAnnounceAddrEvent](),
		ShouldRequestPeers: topic.NewTopic[ShouldRequestPeersEvent](),
		ValidatedHead:      topic.NewTopic[ValidatedHeadEvent](),
		ValidatedTx:        topic.NewTopic[ValidatedTxEvent](),
	}
}
