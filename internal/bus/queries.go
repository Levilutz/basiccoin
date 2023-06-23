package bus

import "github.com/levilutz/basiccoin/pkg/core"

// A query for the balance of a PublicKeyHash.
type PkhBalanceQuery struct {
	Ret             chan map[core.HashT]uint64
	PublicKeyHashes []core.HashT
}

// A query for the current utxos controlled by a PublicKeyHash.
type PkhUtxosQuery struct {
	Ret           chan []core.Utxo
	PublicKeyHash core.HashT
}
