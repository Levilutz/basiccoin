package db

// Unspent transaction output.
type UTxO struct {
	TxId HashT
	Ind  int
}

// Auxiliary computed transaction information.
type TxAux struct {
	RefOutputs []TxOut
	Surplus    int
	VSize      int
}

// Total state of the blockchain and mempool.
type State struct {
	// currentHead    HashT
	// ledgerTxBlocks map[HashT]HashT
	ledger  map[HashT]BlockHeader
	txs     map[HashT]Tx
	uTxOs   map[UTxO]struct{}
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
