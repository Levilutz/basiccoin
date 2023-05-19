package events

// True Events

// Command to terminate the connection.
type ShouldEndPeerEvent struct {
	// Whether we should notify the main bus of the closure.
	NotifyMainBus bool
}

// Inform the peer of a Block with its Merkle tree and Txs.
type BlockDataPeerEvent struct {
	// Block, Merkle, Txs
}

// Inform the peer of our mempool's Txs.
type MempoolDataPeerEvent struct {
	// Mempool, Txs
}

// Inform the peer of other peers.
type PeersDataPeerEvent struct {
	PeerAddrs []string
}

// Query Events

// Retrieve other peers from the peer.
// Responds on MainBus:PeersReceived.
type PeersWantedPeerEvent struct {
}

// Retrieve a Block with its Merkle tree and Txs from the peer.
// Responds on MainBus:BlockReceived.
type BlockWantedPeerEvent struct {
}

// Retrieve the peer's mempool.
// Responds on MainBus:TxsReceived.
type MempoolWantedPeerEvent struct {
}
