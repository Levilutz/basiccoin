package db

// Unspent transaction output.
type UTXO struct {
	TxId HashT
	Ind  uint32
}

func (utxo UTXO) Hash() HashT {
	return DHashItems(utxo.TxId, utxo.Ind)
}

// Total state of the blockchain and mempool.
type State struct {
	// head    HashT
	// mempool map[HashT]struct{}
	// utxos   map[HashT]struct{}
}
