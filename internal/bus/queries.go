package bus

import "github.com/levilutz/basiccoin/pkg/core"

// A query for the balance of a PublicKeyHash.
type PkhBalanceQuery struct {
	Ret             chan map[core.HashT]uint64
	PublicKeyHashes []core.HashT
}

// A query for the current utxos controlled by a PublicKeyHash.
type PkhUtxosQuery struct {
	Ret             chan map[core.Utxo]core.HashT
	PublicKeyHashes []core.HashT
}

// A query for the number of confirmations for each given tx.
// If any of the given txs aren't known, they're not returned in the output map.
type TxConfirmsQuery struct {
	Ret   chan map[core.HashT]uint64
	TxIds []core.HashT
}
