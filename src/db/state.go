package db

import "github.com/levilutz/basiccoin/src/util"

// Unspent transaction output.
type Utxo struct {
	TxId HashT
	Ind  uint32
}

// State at a blockchain node.
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
func (s *State) Rewind() error {
	return nil
}

// Verify whether a state should be allowed to advance to this block.
func (s *State) CanAdvance(next Block) error {
	return nil
}

// Advance a state to the next block, does not verify.
func (s *State) Advance(next Block) error {
	return nil
}
