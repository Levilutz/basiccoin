package db

import (
	"fmt"
	"sort"

	"github.com/levilutz/basiccoin/src/util"
)

// State at a blockchain node. Responsible for preventing double-spends.
// Meant to only be accessed synchronously by a single thread.
type State struct {
	head         HashT
	mempool      *util.Set[HashT]
	utxos        *util.Set[Utxo]
	inv          InvReader
	mempoolRates map[HashT]float64
}

func NewState(inv InvReader) *State {
	return &State{
		head:         HashTZero,
		mempool:      util.NewSet[HashT](),
		utxos:        util.NewSet[Utxo](),
		inv:          inv,
		mempoolRates: make(map[HashT]float64),
	}
}

// Copy a state.
func (s *State) Copy() *State {
	return &State{
		head:         s.head,
		mempool:      s.mempool.Copy(),
		utxos:        s.utxos.Copy(),
		inv:          s.inv,
		mempoolRates: util.CopyMap(s.mempoolRates),
	}
}

func (s *State) GetHead() HashT {
	return s.head
}

// Rewind a state to its parent block.
// If this fails state will be corrupted, so copy before if necessary.
func (s *State) Rewind() error {
	if s.head == HashTZero {
		return fmt.Errorf("cannot rewind - at root")
	}
	rBlock := s.inv.GetBlock(s.head)
	rTxs := s.inv.GetMerkleTxs(rBlock.MerkleRoot)
	for _, tx := range rTxs {
		txId := tx.Hash()
		// Return tx back to mempool
		s.mempool.Add(txId)
		s.mempoolRates[txId] = tx.Rate()
		// Return the tx inputs
		for _, utxo := range tx.GetUtxos() {
			s.utxos.Add(utxo)
		}
		// Remove the tx outputs from the utxo set
		for i := range tx.Outputs {
			if !s.utxos.Remove(Utxo{TxId: txId, Ind: uint64(i)}) {
				return fmt.Errorf("state corrupt - missing utxo %x[%d]", txId, i)
			}
		}
	}
	s.head = rBlock.PrevBlockId
	return nil
}

// Rewind a state until head is the given block.
// If this fails state will be corrupted, so copy before if necessary.
func (s *State) RewindUntil(blockId HashT) {
	depth, ok := s.inv.GetBlockAncestorDepth(s.head, blockId)
	if !ok {
		panic(fmt.Sprintf("head does not have ancestor %x", blockId))
	}
	for i := uint64(0); i < depth; i++ {
		fmt.Println("!!! rewinding!")
		if err := s.Rewind(); err != nil {
			panic(err)
		}
	}
	if s.head != blockId {
		panic(fmt.Sprintf("head is not expected value: %x != %x", s.head, blockId))
	}
}

// Advance a state to a given next block.
// If this fails state will be corrupted, so copy before if necessary.
func (s *State) Advance(nextBlockId HashT) error {
	if !s.inv.HasBlock(nextBlockId) {
		return fmt.Errorf("cannot advance, block unknown: %x", nextBlockId)
	}
	nBlock := s.inv.GetBlock(nextBlockId)
	nTxs := s.inv.GetMerkleTxs(nBlock.MerkleRoot)
	if nBlock.PrevBlockId != s.head {
		return fmt.Errorf(
			"block not based on this parent: %x != %x", nBlock.PrevBlockId, s.head,
		)
	}
	for _, tx := range nTxs {
		txId := tx.Hash()
		err := s.VerifyTxIncludable(txId)
		if err != nil {
			return fmt.Errorf("tx not includable: %s", err.Error())
		}
		// Verify above min block height
		height := s.inv.GetBlockHeight(nextBlockId)
		if height < tx.MinBlock {
			return fmt.Errorf("tx cannot be included in block - too low")
		}
		// Remove tx from mempool
		if !s.mempool.Remove(txId) {
			return fmt.Errorf("state corrupt - missing tx %x", txId)
		}
		delete(s.mempoolRates, txId)
		// Consume the tx inputs
		for _, utxo := range tx.GetUtxos() {
			if !s.utxos.Remove(utxo) {
				return fmt.Errorf("tx input not available %x[%d]", utxo.TxId, utxo.Ind)
			}
		}
		// Add the tx outputs
		for i := range tx.Outputs {
			s.utxos.Add(Utxo{TxId: txId, Ind: uint64(i)})
		}
	}
	s.head = nextBlockId
	return nil
}

// Check whether a tx can be included in a new block based on this head.
func (s *State) VerifyTxIncludable(txId HashT) error {
	if !s.inv.HasTx(txId) {
		return ErrEntityUnknown
	}
	tx := s.inv.GetTx(txId)
	height := s.inv.GetBlockHeight(s.head)
	if height+1 < tx.MinBlock {
		return fmt.Errorf("tx cannot be included in block - too low")
	}
	if !s.mempool.Includes(txId) {
		return fmt.Errorf("tx does not exist in mempool")
	}
	// Verify each tx input's claimed utxo is available
	// This guards against double-spends
	for _, utxo := range tx.GetUtxos() {
		if !s.utxos.Includes(utxo) {
			return fmt.Errorf("tx input not available %x[%d]", utxo.TxId, utxo.Ind)
		}
	}
	return nil
}

// Get includable mempool txs sorted be fee rate, descending.
func (s *State) GetSortedIncludableMempool() []HashT {
	mem := s.mempool.Copy()
	mem.Filter(func(txId HashT) bool {
		return s.VerifyTxIncludable(txId) == nil && s.inv.GetTx(txId).HasSurplus()
	})
	memL := mem.ToList()
	sort.Slice(memL, func(i, j int) bool {
		// > instead of < because we want descending.
		return s.mempoolRates[memL[i]] > s.mempoolRates[memL[j]]
	})
	return memL
}

// Add a tx to the mempool.
func (s *State) AddMempoolTx(txId HashT) {
	tx := s.inv.GetTx(txId)
	s.mempool.Add(txId)
	s.mempoolRates[txId] = tx.Rate()
}
