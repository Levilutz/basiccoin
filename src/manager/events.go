package manager

import "github.com/levilutz/basiccoin/src/db"

// A Peer is closing its connection, remove from table.
type peerClosingEvent struct {
	RuntimeID string
}

// Save the newly-discovered peer addresses.
type peersReceivedEvent struct {
	PeerAddrs []string
}

// A candidate (unverified) set of blocks to upgrade the ledger to, with needed data.
type inboundSyncEvent struct {
	Head    db.HashT
	Blocks  []db.Block
	Merkles []db.MerkleNode
	Txs     []db.Tx
}

// Retrieve our currently known peers.
type peersWantedEvent struct {
	PeerRuntimeID string
}
