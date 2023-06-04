package manager

import "github.com/levilutz/basiccoin/src/db"

// A Peer is closing its connection, remove from table.
type peerClosingEvent struct {
	runtimeID string
}

// Save the newly-discovered peer addresses.
type peersReceivedEvent struct {
	peerAddrs []string
}

// A candidate (unverified) set of blocks to upgrade the ledger to, with needed data.
type inboundSyncEvent struct {
	head    db.HashT
	blocks  []db.Block
	merkles []db.MerkleNode
	txs     []db.Tx
}

// Retrieve our currently known peers.
type peersWantedEvent struct {
	peerRuntimeID string
}

// Store a new tx.
type newTxEvent struct {
	tx db.Tx
}

// Get the balance of a public key hash.
type balanceQuery struct {
	rCh           chan<- uint64
	publicKeyHash db.HashT
}
