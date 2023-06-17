package peer

import "github.com/levilutz/basiccoin/src/db"

// Command to terminate the connection.
type shouldEndEvent struct{}

// Inform this peer of our head, sync if desired.
type syncHeadEvent struct {
	head db.HashT2
}

// Inform the peer of other peers.
type peersDataEvent struct {
	addrs []string
}

// Retrieve other peers from the peer.
type peersWantedEvent struct{}

// Inform the peer of a new tx.
type sendTxEvent struct {
	txId db.HashT2
}
