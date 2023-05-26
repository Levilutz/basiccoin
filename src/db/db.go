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
	// Mutable
	head     HashT
	mempool  map[HashT]struct{}
	curUtxos map[HashT]struct{}
	// Write-once maps
	ledger    map[HashT]BlockHeader
	blockHs   map[HashT]uint32
	merkles   map[HashT]MerkleNode
	txs       map[HashT]Tx
	usedUtxos map[HashT]HashT
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

// Whether the given UTXO exists at the conclusion of the given block id.
// func (s *State) UtxoExists(utxo UTXO, blockId HashT) bool {
// 	usedBlock, ok := s.utxos[utxo.Hash()]
// 	if !ok {
// 		return false
// 	}
// 	if usedBlock == BlankHash && blockId == s.head {
// 		return true
// 	}
// 	// Whether usedHeight > blockHeight
// 	return s.blockHs[usedBlock] > s.blockHs[blockId]
// }
