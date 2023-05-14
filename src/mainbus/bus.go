package mainbus

type MainBus struct {
	PeerClosings chan string
	Events       chan MainBusEvent
	UrgentEvents chan MainBusEvent
}

func NewMainBus(bufferSize int) *MainBus {
	return &MainBus{
		PeerClosings: make(chan string, bufferSize),
		Events:       make(chan MainBusEvent, bufferSize),
		UrgentEvents: make(chan MainBusEvent, bufferSize),
	}
}

type MainBusEvent struct {
	// True Events
	BlockReceived  *BlockReceivedEvent
	MerkleReceived *MerkleReceivedEvent
	TxsReceived    *TxsReceivedEvent
	NewPeers       *NewPeersEvent
	// Query Events
	PeersWanted  *PeersWantedEvent
	BlockWanted  *BlockWantedEvent
	MerkleWanted *MerkleWantedEvent
	TxsWanted    *TxsWantedEvent
}

type BlockReceivedEvent struct {
}

type MerkleReceivedEvent struct {
}

type TxsReceivedEvent struct {
}

type NewPeersEvent struct {
}

type PeersWantedEvent struct {
}

type BlockWantedEvent struct {
}

type MerkleWantedEvent struct {
}

type TxsWantedEvent struct {
}
