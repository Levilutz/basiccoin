package db

import "github.com/levilutz/basiccoin/src/util"

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
		for i, _ := range tx.Outputs {
			if !s.Utxos.Remove(Utxo{TxId: txId, Ind: uint32(i)}) {
				return err
			}
		}
	}
	s.Head = rBlock.PrevBlockId
	return nil
}

// Verify whether a state should be allowed to advance to this block.
func (s *State) CanAdvance(next Block) error {
	return nil
}

// Advance a state to the next block, does not verify.
// If this fails state will be corrupted, so copy before if necessary.
func (s *State) Advance(next Block) error {
	return nil
}
