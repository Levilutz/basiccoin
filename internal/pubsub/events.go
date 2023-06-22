package pubsub

import "github.com/levilutz/basiccoin/pkg/core"

// When we have a new potential head for the chain to validate.
type CandidateHeadEvent struct {
	Head    core.HashT
	Blocks  []core.Block
	Merkles []core.MerkleNode
	Txs     []core.Tx
}

// When we have a new potential tx for the chain to validate.
type CandidateTxEvent struct {
	Tx core.Tx
}

// Emitted alongside ValidatedHeatEvent, if miners are running.
// Informs the miners of what set of Txs is most profitable to include now.
type MinerTargetEvent struct {
	Head   core.HashT
	Target core.HashT
	TxIds  []core.HashT
}

// A peer has announced its listen address.
type PeerAnnouncedAddrEvent struct {
	PeerRuntimeId string
	Addr          string
}

// Emitted by a peer as it closes.
type PeerClosingEvent struct {
	PeerRuntimeId string
}

// We have received the addresses of other peers.
type PeersReceivedEvent struct {
	PeerAddrs map[string]string
}

// The specified peer has requested a list of our peers.
type PeersRequestedEvent struct {
	PeerRuntimeId string
}

// The specified components should print an update.
type PrintUpdateEvent struct {
	Peer        bool
	PeerFactory bool
}

// Send the given peers address list to the specified peer.
type SendPeersEvent struct {
	TargetRuntimeId string
	PeerAddrs       map[string]string
}

// We should announce our address to a peer.
type ShouldAnnounceAddrEvent struct {
	TargetRuntimeId string
	Addr            string
}

// We should request the given peer id for their peers.
type ShouldRequestPeersEvent struct {
	TargetRuntimeId string
}

// Emitted by the chain when we advance to a new head.
type ValidatedHeadEvent struct {
	Head core.HashT
}

// Emitted by the chain when we validate a new tx.
type ValidatedTxEvent struct {
	TxId core.HashT
}
