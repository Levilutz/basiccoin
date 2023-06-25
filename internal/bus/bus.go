package bus

import "github.com/levilutz/basiccoin/pkg/topic"

// The set of pub sub topics any component needs.
type Bus struct {
	// Events
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
	// Commands
	Terminate *topic.Topic[TerminateCommand]
	// Queries
	HeadHeight      *topic.Topic[HeadHeightQuery]
	PkhBalance      *topic.Topic[PkhBalanceQuery]
	PkhUtxos        *topic.Topic[PkhUtxosQuery]
	TxConfirms      *topic.Topic[TxConfirmsQuery]
	TxIncludedBlock *topic.Topic[TxIncludedBlockQuery]
}

func NewBus() *Bus {
	return &Bus{
		// Events
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
		// Commands
		Terminate: topic.NewTopic[TerminateCommand](),
		// Queries
		HeadHeight:      topic.NewTopic[HeadHeightQuery](),
		PkhBalance:      topic.NewTopic[PkhBalanceQuery](),
		PkhUtxos:        topic.NewTopic[PkhUtxosQuery](),
		TxConfirms:      topic.NewTopic[TxConfirmsQuery](),
		TxIncludedBlock: topic.NewTopic[TxIncludedBlockQuery](),
	}
}
