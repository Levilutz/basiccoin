package db

import (
	"fmt"

	"github.com/levilutz/basiccoin/src/util"
)

// Unspent transaction output.
type Utxo struct {
	TxId HashT
	Ind  uint32
}

func UtxoFromInput(txi TxIn) Utxo {
	return Utxo{
		TxId: txi.OriginTxId,
		Ind:  txi.OriginTxOutInd,
	}
}

// State at a blockchain node.
// Meant to only be accessed synchronously by a single thread.
type State struct {
	Head    HashT
	Mempool *util.Set[HashT]
	Utxos   *util.Set[Utxo]
	inv     InvReader
}

// Copy a state.
func (s *State) Copy() *State {
	return &State{
		Head:    s.Head,
		Mempool: s.Mempool.Copy(),
		Utxos:   s.Utxos.Copy(),
		inv:     s.inv,
	}
}

// Rewind a state to its parent block.
// If this fails state will be corrupted, so copy before if necessary.
func (s *State) Rewind() error {
	rBlock, _, rTxs, err := s.inv.LoadFullBlock(s.Head)
	if err != nil {
		return err
	}
	for txId, tx := range rTxs {
		// Return tx back to mempool
		s.Mempool.Add(txId)
		// Return the tx inputs
		for _, txi := range tx.Inputs {
			s.Utxos.Add(UtxoFromInput(txi))
		}
		// Remove the tx outputs from the utxo set
		for i := range tx.Outputs {
			if !s.Utxos.Remove(Utxo{TxId: txId, Ind: uint32(i)}) {
				return fmt.Errorf("state corrupt - missing utxo %x[%d]", txId, i)
			}
		}
	}
	s.Head = rBlock.PrevBlockId
	return nil
}

// Rewind a state until head is the given block.
// If this fails state will be corrupted, so copy before if necessary.
func (s *State) RewindUntil(blockId HashT) error {
	depth, err := s.inv.AncestorDepth(s.Head, blockId)
	if err != nil {
		return err
	}
	for i := uint32(0); i < depth; i++ {
		if err := s.Rewind(); err != nil {
			return err
		}
	}
	if s.Head != blockId {
		return fmt.Errorf("head is not expected value: %x != %x", s.Head, blockId)
	}
	return nil
}

// Advance a state to a given next block, does not verify much.
// If this fails state will be corrupted, so copy before if necessary.
func (s *State) Advance(nextBlockId HashT) error {
	nBlock, _, nTxs, err := s.inv.LoadFullBlock(nextBlockId)
	if err != nil {
		return err
	}
	if nBlock.PrevBlockId != s.Head {
		return fmt.Errorf("block not based on this parent")
	}
	for txId, tx := range nTxs {
		// Remove tx from mempool
		if !s.Mempool.Remove(txId) {
			return fmt.Errorf("state corrupt - missing tx %x", txId)
		}
		// Consume the tx inputs
		for _, txi := range tx.Inputs {
			if !s.Utxos.Remove(UtxoFromInput(txi)) {
				return fmt.Errorf(
					"tx input not available %x[%d]", txi.OriginTxId, txi.OriginTxOutInd,
				)
			}
		}
		// Add the tx outputs
		for i := range tx.Outputs {
			s.Utxos.Add(Utxo{TxId: txId, Ind: uint32(i)})
		}
	}
	s.Head = nextBlockId
	return nil
}

// Advance the state through the given next blocks, does not verify much.
// If this fails state will be corrupted, so copy before if necessary.
func (s *State) AdvanceMany(nextBlockIds []HashT) error {
	for _, nextBlockId := range nextBlockIds {
		if err := s.Advance(nextBlockId); err != nil {
			return err
		}
	}
	return nil
}
