package main

import (
	"github.com/levilutz/basiccoin/src/kern"
)

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
	head    kern.HashT
	blocks  []kern.Block
	merkles []kern.MerkleNode
	txs     []kern.Tx
}

// Retrieve our currently known peers.
type peersWantedEvent struct {
	peerRuntimeID string
}

// Store a new tx.
type newTxEvent struct {
	tx kern.Tx
}

// Get the balakern.HashT public key hash.
type balanceQuery struct {
	rCh           chan<- uint64
	publicKeyHash kern.HashT
}

// Get the utxokern.HashTble to a given public key hash.
type utxosQuery struct {
	rCh           chan<- []kern.Utxo
	publicKeyHash kern.HashT
}

// Store a new tx, respond with success.
type newTxQuery struct {
	rCh chan<- error
	tx  kern.Tx
}
