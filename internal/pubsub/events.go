package pubsub

import "github.com/levilutz/basiccoin/pkg/core"

// Emitted by a peer as it closes.
type PeerClosingEvent struct {
	PeerRuntimeId string
}

// We should request the given peer id for their peers.
type ShouldRequestPeersEvent struct {
	PeerRuntimeId string
}

// Emitted by the chain when we advance to a new head.
type ValidatedHeadEvent struct {
	Head core.HashT
}
