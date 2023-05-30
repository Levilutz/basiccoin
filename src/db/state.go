package db

import (
	"fmt"
	"sort"

	"github.com/levilutz/basiccoin/src/util"
)

// Unspent transaction output.
type Utxo struct {
	TxId HashT
	Ind  uint64
}

func UtxoFromInput(txi TxIn) Utxo {
	return Utxo{
		TxId: txi.OriginTxId,
		Ind:  txi.OriginTxOutInd,
	}
}

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
		return fmt.Errorf("cannot rewind - at origin")
	}
	rBlock, ok := s.inv.LoadBlock(s.head)
	if !ok {
		return ErrEntityUnknown
	}
	rTxs, err := s.inv.LoadMerkleTxs(rBlock.MerkleRoot)
	if err != nil {
		return err
	}
	for _, tx := range rTxs {
		txId := tx.Hash()
		// Return tx back to mempool
		s.mempool.Add(txId)
		s.mempoolRates[txId] = tx.Rate()
		// Return the tx inputs
		for _, txi := range tx.Inputs {
			s.utxos.Add(UtxoFromInput(txi))
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
func (s *State) RewindUntil(blockId HashT) error {
	depth, err := s.inv.AncestorDepth(s.head, blockId)
	if err != nil {
		return err
	}
	for i := uint64(0); i < depth; i++ {
		if err := s.Rewind(); err != nil {
			return err
		}
	}
	if s.head != blockId {
		return fmt.Errorf("head is not expected value: %x != %x", s.head, blockId)
	}
	return nil
}

// Advance a state to a given next block.
// If this fails state will be corrupted, so copy before if necessary.
func (s *State) Advance(nextBlockId HashT) error {
	nBlock, ok := s.inv.LoadBlock(nextBlockId)
	if !ok {
		return fmt.Errorf("cannot advance, block unknown: %x", nextBlockId)
	}
	nTxs, err := s.inv.LoadMerkleTxs(nBlock.MerkleRoot)
	if err != nil {
		return fmt.Errorf("failed to load merkle txs: %s", err.Error())
	}
	if nBlock.PrevBlockId != s.head {
		return fmt.Errorf("block not based on this parent")
	}
	for _, tx := range nTxs {
		txId := tx.Hash()
		err := s.VerifyTxIncludable(txId)
		if err != nil {
			return fmt.Errorf("tx not includable: %s", err.Error())
		}
		// Verify above min block height
		height, err := s.inv.GetBlockHeight(nextBlockId)
		if err != nil {
			return fmt.Errorf("failed to retrieve next block height: %s", err.Error())
		}
		if height < tx.MinBlock {
			return fmt.Errorf("tx cannot be included in block - too low")
		}
		// Remove tx from mempool
		if !s.mempool.Remove(txId) {
			return fmt.Errorf("state corrupt - missing tx %x", txId)
		}
		delete(s.mempoolRates, txId)
		// Consume the tx inputs
		for _, txi := range tx.Inputs {
			if !s.utxos.Remove(UtxoFromInput(txi)) {
				return fmt.Errorf(
					"tx input not available %x[%d]", txi.OriginTxId, txi.OriginTxOutInd,
				)
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

// Advance the state through the given next blocks.
// If this fails state will be corrupted, so copy before if necessary.
func (s *State) AdvanceMany(nextBlockIds []HashT) error {
	for _, nextBlockId := range nextBlockIds {
		if err := s.Advance(nextBlockId); err != nil {
			return err
		}
	}
	return nil
}

// Check whether a tx can be included in a new block based on this head.
func (s *State) VerifyTxIncludable(txId HashT) error {
	tx, ok := s.inv.LoadTx(txId)
	if !ok {
		return ErrEntityUnknown
	}
	height, err := s.inv.GetBlockHeight(s.head)
	if err != nil {
		return err
	}
	if height+1 < tx.MinBlock {
		return fmt.Errorf("tx cannot be included in block - too low")
	}
	if !s.mempool.Includes(txId) {
		return fmt.Errorf("tx does not exist in mempool")
	}
	// Verify each tx input's claimed utxo is available
	// This is the primary guard against double-spends
	for _, txi := range tx.Inputs {
		if !s.utxos.Includes(UtxoFromInput(txi)) {
			return fmt.Errorf(
				"tx input not available %x[%d]", txi.OriginTxId, txi.OriginTxOutInd,
			)
		}
	}
	return nil
}

// Get includable mempool txs sorted be fee rate, descending.
func (s *State) GetSortedIncludableMempool() []HashT {
	mem := s.mempool.Copy()
	mem.Filter(func(key HashT) bool {
		return s.VerifyTxIncludable(key) == nil
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
	tx, ok := s.inv.LoadTx(txId)
	if !ok {
		panic("tx should exist")
	}
	s.mempool.Add(txId)
	s.mempoolRates[txId] = tx.Rate()
}
