package peer

type PeerBus struct {
	PeerRuntimeID string
	Events        chan PeerBusEvent
	UrgentEvents  chan PeerBusEvent
}

func NewPeerBus(peerRuntimeID string, bufferSize int) *PeerBus {
	return &PeerBus{
		PeerRuntimeID: peerRuntimeID,
		Events:        make(chan PeerBusEvent, bufferSize),
		UrgentEvents:  make(chan PeerBusEvent, bufferSize),
	}
}

type PeerBusEvent struct {
	// True Events
	ShouldEnd  *ShouldEndEvent
	BlockData  *BlockDataEvent
	MerkleData *MerkleDataEvent
	TxsData    *TxsDataEvent
	// Query Events
	PeersWanted  *PeersWantedEvent
	BlockWanted  *BlockWantedEvent
	MerkleWanted *MerkleWantedEvent
	TxsWanted    *TxsWantedEvent
}

// True Events

type ShouldEndEvent struct {
	SendClose     bool
	NotifyMainBus bool
}

type BlockDataEvent struct {
}

type MerkleDataEvent struct {
}

type TxsDataEvent struct {
}

// Query Events

type PeersWantedEvent struct {
}

type BlockWantedEvent struct {
}

type MerkleWantedEvent struct {
}

type TxsWantedEvent struct {
}
