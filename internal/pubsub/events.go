package pubsub

import "github.com/levilutz/basiccoin/pkg/core"

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

// Send the given peers address list to the specified peer.
type SendPeersEvent struct {
	PeerRuntimeId string
	PeerAddrs     map[string]string
}

// We should request the given peer id for their peers.
type ShouldRequestPeersEvent struct {
	PeerRuntimeId string
}

// Emitted by the chain when we advance to a new head.
type ValidatedHeadEvent struct {
	Head core.HashT
}
