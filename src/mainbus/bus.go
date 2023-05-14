package mainbus

type MainBus struct {
	Events chan MainBusEvent
}

func NewMainBus(bufferSize int) *MainBus {
	return &MainBus{
		Events: make(chan MainBusEvent, bufferSize),
	}
}

type MainBusEvent struct {
	// True Events
	PeerClosing   *PeerClosingEvent
	BlockReceived *BlockReceivedEvent
	TxsReceived   *TxsReceivedEvent
	PeersReceived *PeersReceivedEvent
	// Query Events
	PeersWanted   *PeersWantedEvent
	BlockWanted   *BlockWantedEvent
	MempoolWanted *MempoolWantedEvent
}

// A Peer is closing its connection, remove from table.
type PeerClosingEvent struct {
	RuntimeID string
}

// Save the newly-received block.
type BlockReceivedEvent struct {
	// Block, Merkle, Txs
}

// Save the newly-received txs.
type TxsReceivedEvent struct {
}

// Save the newly-discovered peer addresses.
type PeersReceivedEvent struct {
	PeerAddrs []string
}

// Query Events

// Retrieve our currently known peers.
// Responds on PeerBus:PeerData.
type PeersWantedEvent struct {
	PeerRuntimeID string
}

// Retrieve a known Block with its Merkle tree and Txs.
// Responds on PeerBus:BlockData.
type BlockWantedEvent struct {
}

// Retrieve our mempool's Txs.
// Responds on PeerBus:MempoolDataEvent
type MempoolWantedEvent struct {
}
