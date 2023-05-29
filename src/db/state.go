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

// Verify whether a state should be allowed to advance to the given next block.
func (s *State) ShouldAdvance(nextBlockId HashT) error {
	return nil
}

// Advance a state to a given next block, does not verify.
// If this fails state will be corrupted, so copy before if necessary.
func (s *State) Advance(nextBlockId HashT) error {
	nBlock, _, nTxs, err := s.inv.LoadFullBlock(nextBlockId)
	if nBlock.PrevBlockId != s.Head {
		return fmt.Errorf("block not based on this parent")
	}
	if err != nil {
		return err
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
