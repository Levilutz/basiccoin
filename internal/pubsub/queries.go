package pubsub

import "github.com/levilutz/basiccoin/pkg/core"

// A query for the balance of a particular PublicKeyHash.
type PkhBalanceQuery struct {
	Ret           chan uint64
	PublicKeyHash core.HashT
}
