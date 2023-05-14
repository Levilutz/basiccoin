package peer

type PeerEvent struct {
	// True Events
	ShouldEnd   *ShouldEndEvent
	BlockData   *BlockDataEvent
	MempoolData *MempoolDataEvent
	PeersData   *PeersDataEvent
	// Query Events
	PeersWanted   *PeersWantedEvent
	BlockWanted   *BlockWantedEvent
	MempoolWanted *MempoolWantedEvent
}

// True Events

// Command to terminate the connection.
type ShouldEndEvent struct {
	SendClose     bool
	NotifyMainBus bool
}

// Inform the peer of a Block with its Merkle tree and Txs.
type BlockDataEvent struct {
	// Block, Merkle, Txs
}

// Inform the peer of our mempool's Txs.
type MempoolDataEvent struct {
	// Mempool, Txs
}

// Inform the peer of other peers.
type PeersDataEvent struct {
}

// Query Events

// Retrieve other peers from the peer.
// Responds on MainBus:PeersReceived.
type PeersWantedEvent struct {
}

// Retrieve a Block with its Merkle tree and Txs from the peer.
// Responds on MainBus:BlockReceived.
type BlockWantedEvent struct {
}

// Retrieve the peer's mempool.
// Responds on MainBus:TxsReceived.
type MempoolWantedEvent struct {
}
