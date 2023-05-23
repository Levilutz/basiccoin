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
	head    HashT
	ledger  map[HashT]BlockHeader
	merkles map[HashT]MerkleNode
	txs     map[HashT]Tx
	utxos   map[HashT]struct{}
	mempool map[HashT]struct{}
}

// Add a tx to the mempool, fails if it exists in ledger.
func (s *State) AddMempoolTx(hash HashT, tx Tx) error {
	_, exists := s.txs[hash]
	if !exists {
		s.txs[hash] = tx
		s.mempool[hash] = struct{}{}
	}
	return nil
}

func (s *State) UtxoExists(utxo UTXO, blockId HashT) bool {
	if blockId == s.head {
		_, ok := s.utxos[utxo.Hash()]
		return ok
	}
	// TODO: Rewind
	return false
}
