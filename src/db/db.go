package db

import "fmt"

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

// Total state of the blockchain and mempool
type State struct {
	CurrentHead    HashT
	Ledger         map[HashT]Block
	LedgerTxs      map[HashT]Tx
	LedgerTxBlocks map[HashT]HashT
	UTxOs          map[UTxO]struct{}
	Mempool        map[HashT]Tx
}

// Add a tx to the mempool, fails if it exists in ledger.
func (s *State) AddMempoolTx(hash HashT, tx Tx) error {
	_, exists := s.LedgerTxs[hash]
	if exists {
		return fmt.Errorf("cannot add to mempool, exists in ledger: %s", hash)
	}
	_, exists = s.Mempool[hash]
	if !exists {
		s.Mempool[hash] = tx
	}
	return nil
}
