package bus

import "github.com/levilutz/basiccoin/pkg/core"

// A query for the current height of the chain head.
type HeadHeightQuery struct {
	Ret chan uint64
}

// A query for the balance of a PublicKeyHash.
type PkhBalanceQuery struct {
	Ret             chan map[core.HashT]uint64
	PublicKeyHashes []core.HashT
}

// A query for the current utxos controlled by a PublicKeyHash.
// Optionally, exclude utxos that are spent by any txs in the mempool.
type PkhUtxosQuery struct {
	Ret             chan map[core.Utxo]core.HashT
	PublicKeyHashes []core.HashT
	ExcludeMempool  bool
}

// A query for the number of confirmations for each given tx.
// If any of the given txs aren't known, they're not returned in the output map.
type TxConfirmsQuery struct {
	Ret   chan map[core.HashT]uint64
	TxIds []core.HashT
}

// A query for the block id a tx was included in, for each given tx.
type TxIncludedBlockQuery struct {
	Ret   chan map[core.HashT]core.HashT
	TxIds []core.HashT
}
