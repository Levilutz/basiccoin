package bus

import "github.com/levilutz/basiccoin/pkg/core"

// A query for the balance of a PublicKeyHash.
type PkhBalanceQuery struct {
	Ret           chan uint64
	PublicKeyHash core.HashT
}

// A query for the current utxos controlled by a PublicKeyHash.
type PkhUtxosQuery struct {
	Ret           chan []core.Utxo
	PublicKeyHash core.HashT
}
