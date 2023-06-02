package events

import "github.com/levilutz/basiccoin/src/db"

// True Events

// A Peer is closing its connection, remove from table.
type PeerClosingMainEvent struct {
	// ID of the Peer that's closing.
	RuntimeID string
}

// Save the newly-received block.
type BlockReceivedMainEvent struct {
	// Block, Merkle, Txs
}

// Save the newly-received txs.
type TxsReceivedMainEvent struct {
}

// Save the newly-discovered peer addresses.
type PeersReceivedMainEvent struct {
	PeerAddrs []string
}

// A candidate (unverified) set of blocks to upgrade the ledger to, with needed data.
type InboundSyncMainEvent struct {
	Head    db.HashT
	Blocks  []db.Block
	Merkles []db.MerkleNode
	Txs     []db.Tx
}

// Query Events

// Retrieve our currently known peers.
// Responds on PeerBus:PeersData.
type PeersWantedMainEvent struct {
	// ID of Peer that wants response.
	PeerRuntimeID string
}

// Retrieve a known Block with its Merkle tree and Txs.
// Responds on PeerBus:BlockData.
type BlockWantedMainEvent struct {
	// ID of Peer that wants response.
	PeerRuntimeID string
}

// Retrieve our mempool's Txs.
// Responds on PeerBus:MempoolData.
type MempoolWantedMainEvent struct {
	// ID of Peer that wants response.
	PeerRuntimeID string
}
