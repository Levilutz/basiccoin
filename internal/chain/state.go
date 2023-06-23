package chain

import (
	"fmt"
	"sort"

	"github.com/levilutz/basiccoin/internal/inv"
	"github.com/levilutz/basiccoin/pkg/core"
	"github.com/levilutz/basiccoin/pkg/set"
	"github.com/levilutz/basiccoin/pkg/util"
)

// The state of our local blockchain.
// Should only be accessed by a single thread.
type State struct {
	// The head of our current chain
	head core.HashT

	// The set of txs we have verified but are not included in current chain
	mempool *set.Set[core.HashT]

	// The set of utxos that haven't been spent at the head of our chain
	utxos *set.Set[core.Utxo]

	// A reference to our inventory
	inv inv.InvReader

	// The fee rate for every tx in our mempool
	mempoolRates map[core.HashT]float64

	// For each utxo spent in the mempool, which mempool txIds spend it
	mempoolUtxoSpends map[core.Utxo]*set.Set[core.HashT]

	// The set of utxos controlled by each public key hash with a balance
	pkhUtxos map[core.HashT]*set.Set[core.Utxo]

	// The block id at which each transaction was included
	includedTxBlocks map[core.HashT]core.HashT
}

// Create a new empty state at the chain zero block.
func NewState(inv inv.InvReader) *State {
	return &State{
		head:              core.HashT{},
		mempool:           set.NewSet[core.HashT](),
		utxos:             set.NewSet[core.Utxo](),
		inv:               inv,
		mempoolRates:      make(map[core.HashT]float64),
		mempoolUtxoSpends: make(map[core.Utxo]*set.Set[core.HashT]),
		pkhUtxos:          make(map[core.HashT]*set.Set[core.Utxo]),
		includedTxBlocks:  make(map[core.HashT]core.HashT),
	}
}

// Copy a state.
func (s *State) Copy() *State {
	// Must deep copy pkhUtxos
	newPkhUtxos := make(map[core.HashT]*set.Set[core.Utxo], len(s.pkhUtxos))
	for pkh, utxos := range s.pkhUtxos {
		newPkhUtxos[pkh] = utxos.Copy()
	}
	// Must deep copy mempoolUtxoSpends
	newMempoolUtxoSpends := make(map[core.Utxo]*set.Set[core.HashT], len(s.mempoolUtxoSpends))
	for utxo, txIds := range s.mempoolUtxoSpends {
		newMempoolUtxoSpends[utxo] = txIds.Copy()
	}
	// Shallow copy everything else
	return &State{
		head:              s.head,
		mempool:           s.mempool.Copy(),
		utxos:             s.utxos.Copy(),
		inv:               s.inv,
		mempoolRates:      util.CopyMap(s.mempoolRates),
		mempoolUtxoSpends: newMempoolUtxoSpends,
		pkhUtxos:          newPkhUtxos,
		includedTxBlocks:  util.CopyMap(s.includedTxBlocks),
	}
}

// Rewind a state to its parent block.
// If this fails state will be corrupted, so copy before if necessary.
func (s *State) Rewind() {
	if s.head.EqZero() {
		panic("cannot rewind - at root")
	}
	rBlock := s.inv.GetBlock(s.head)
	rTxs := s.inv.GetMerkleTxs(rBlock.MerkleRoot)
	for _, tx := range rTxs {
		txId := tx.Hash()
		// Return tx back to mempool
		s.mempool.Add(txId)
		s.mempoolRates[txId] = tx.Rate()
		for _, utxo := range tx.GetConsumedUtxos() {
			if _, ok := s.mempoolUtxoSpends[utxo]; !ok {
				s.mempoolUtxoSpends[utxo] = set.NewSet[core.HashT]()
			}
			s.mempoolUtxoSpends[utxo].Add(txId)
		}
		// Return the tx inputs
		for _, utxo := range tx.GetConsumedUtxos() {
			s.utxos.Add(utxo)
			if s.pkhUtxos != nil {
				txo := s.inv.GetTxOut(utxo.TxId, utxo.Ind)
				s.creditBalance(txo.PublicKeyHash, utxo)
			}
		}
		// Remove the tx outputs from the utxo set
		for i, txo := range tx.Outputs {
			if !s.utxos.Remove(core.Utxo{TxId: txId, Ind: uint64(i), Value: txo.Value}) {
				panic(fmt.Sprintf("state corrupt - missing utxo %s[%d]", txId, i))
			}
			if s.pkhUtxos != nil {
				s.debitBalance(txo.PublicKeyHash, core.Utxo{
					TxId:  txId,
					Ind:   uint64(i),
					Value: txo.Value,
				})
			}
		}
		// Remove the tx included block (and continue to verify it existed)
		existingBlockId, ok := s.includedTxBlocks[txId]
		if !ok || existingBlockId != s.head {
			panic(fmt.Sprintf("state corrupt - missing/wrong tx block %s", txId))
		}
		delete(s.includedTxBlocks, txId)
	}
	s.head = rBlock.PrevBlockId
}

// Rewind a state until head is the given block.
func (s *State) RewindUntil(blockId core.HashT) {
	depth, ok := s.inv.GetBlockAncestorDepth(s.head, blockId)
	if !ok {
		panic(fmt.Sprintf("head does not have ancestor %s", blockId))
	}
	for i := uint64(0); i < depth; i++ {
		fmt.Println("!!! rewinding!")
		s.Rewind()
	}
	if s.head != blockId {
		panic(fmt.Sprintf("head is not expected value: %s != %s", s.head, blockId))
	}
}

// Advance a state to a given next block.
// If this fails state will be corrupted, so copy before if necessary.
func (s *State) Advance(nextBlockId core.HashT) error {
	if !s.inv.HasBlock(nextBlockId) {
		return fmt.Errorf("cannot advance, block unknown: %s", nextBlockId)
	}
	nBlock := s.inv.GetBlock(nextBlockId)
	nTxs := s.inv.GetMerkleTxs(nBlock.MerkleRoot)
	if nBlock.PrevBlockId != s.head {
		return fmt.Errorf(
			"block not based on this parent: %s != %s", nBlock.PrevBlockId, s.head,
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
			return fmt.Errorf("state corrupt - missing tx %s", txId)
		}
		delete(s.mempoolRates, txId)
		for _, utxo := range tx.GetConsumedUtxos() {
			txIds, ok := s.mempoolUtxoSpends[utxo]
			if !ok {
				return fmt.Errorf("state corrupt - missing mempool utxo set for %v", utxo)
			}
			if !txIds.Remove(txId) {
				return fmt.Errorf("state corrupt - mempool utxo set missing tx %s", txId)
			}
			if txIds.Size() == 0 {
				delete(s.mempoolUtxoSpends, utxo)
			}
		}
		// Consume the tx inputs
		for _, utxo := range tx.GetConsumedUtxos() {
			if !s.utxos.Remove(utxo) {
				return fmt.Errorf("tx input not available %s[%d]", utxo.TxId, utxo.Ind)
			}
			if s.pkhUtxos != nil {
				txo := s.inv.GetTxOut(utxo.TxId, utxo.Ind)
				s.debitBalance(txo.PublicKeyHash, utxo)
			}
		}
		// Add the tx outputs
		for i, txo := range tx.Outputs {
			s.utxos.Add(core.Utxo{TxId: txId, Ind: uint64(i), Value: txo.Value})
			if s.pkhUtxos != nil {
				s.creditBalance(txo.PublicKeyHash, core.Utxo{
					TxId:  txId,
					Ind:   uint64(i),
					Value: txo.Value,
				})
			}
		}
		// Add the tx included block (and continue to verify tx isn't already included)
		existingBlockId, ok := s.includedTxBlocks[txId]
		if ok {
			return fmt.Errorf("tx already included in block %s", existingBlockId)
		}
		s.includedTxBlocks[txId] = nextBlockId
	}
	s.head = nextBlockId
	return nil
}

// Check whether a tx can be included in a new block based on this head.
func (s *State) VerifyTxIncludable(txId core.HashT) error {
	if !s.inv.HasTx(txId) {
		return fmt.Errorf("tx unknown: %s", txId)
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
	for _, utxo := range tx.GetConsumedUtxos() {
		if !s.utxos.Includes(utxo) {
			return fmt.Errorf("tx input not available %s[%d]", utxo.TxId, utxo.Ind)
		}
	}
	return nil
}

// Get includable mempool txs sorted be fee rate, descending.
func (s *State) GetSortedIncludableMempool() []core.HashT {
	mem := s.mempool.Copy()
	mem.Filter(func(txId core.HashT) bool {
		return s.VerifyTxIncludable(txId) == nil && s.inv.GetTx(txId).HasSurplus()
	})
	memL := mem.ToList()
	sort.Slice(memL, func(i, j int) bool {
		// descending
		return s.mempoolRates[memL[i]] > s.mempoolRates[memL[j]]
	})
	return memL
}

// Add a tx to the mempool.
func (s *State) AddMempoolTx(txId core.HashT) {
	tx := s.inv.GetTx(txId)
	s.mempool.Add(txId)
	s.mempoolRates[txId] = tx.Rate()
	for _, utxo := range tx.GetConsumedUtxos() {
		if _, ok := s.mempoolUtxoSpends[utxo]; !ok {
			s.mempoolUtxoSpends[utxo] = set.NewSet[core.HashT]()
		}
		s.mempoolUtxoSpends[utxo].Add(txId)
	}
}

// Add to the utxo set of a public key hash.
func (s *State) creditBalance(publicKeyHash core.HashT, credit core.Utxo) {
	_, ok := s.pkhUtxos[publicKeyHash]
	if !ok {
		s.pkhUtxos[publicKeyHash] = set.NewSet[core.Utxo]()
	}
	s.pkhUtxos[publicKeyHash].Add(credit)
}

// Remove from the utxo set of a public key hash.
func (s *State) debitBalance(publicKeyHash core.HashT, debit core.Utxo) {
	utxos, ok := s.pkhUtxos[publicKeyHash]
	if !ok || !s.pkhUtxos[publicKeyHash].Includes(debit) {
		panic("cannot debit balance, balance does not exist")
	}
	utxos.Remove(debit)
}

// Get the utxos of a public key hash. Optionally, exclude utxos that are spent in the mempool.
func (s *State) GetPkhUtxos(publicKeyHash core.HashT, excludeMempool bool) []core.Utxo {
	utxos, ok := s.pkhUtxos[publicKeyHash]
	if !ok {
		return []core.Utxo{}
	}
	if excludeMempool {
		utxos.Filter(func(utxo core.Utxo) bool {
			_, spentInMempool := s.mempoolUtxoSpends[utxo]
			return !spentInMempool
		})
	}
	return utxos.ToList()
}

// Get the utxos of public key hashes. Optionally, exclude utxos that are spent in the mempool.
func (s *State) GetManyPkhUtxos(publicKeyHashes []core.HashT, excludeMempool bool) map[core.Utxo]core.HashT {
	out := make(map[core.Utxo]core.HashT)
	for _, pkh := range publicKeyHashes {
		utxos := s.GetPkhUtxos(pkh, excludeMempool)
		for _, utxo := range utxos {
			out[utxo] = pkh
		}
	}
	return out
}

func (s *State) GetPkhBalance(publicKeyHash core.HashT) uint64 {
	utxos, ok := s.pkhUtxos[publicKeyHash]
	if !ok {
		return 0
	}
	total := uint64(0)
	for _, utxo := range utxos.ToList() {
		total += utxo.Value
	}
	return total
}

func (s *State) GetManyPkhBalances(publicKeyHashes []core.HashT) map[core.HashT]uint64 {
	out := make(map[core.HashT]uint64, len(publicKeyHashes))
	for _, pkh := range publicKeyHashes {
		out[pkh] = s.GetPkhBalance(pkh)
	}
	return out
}

func (s *State) GetTxConfirms(txIds []core.HashT) map[core.HashT]uint64 {
	out := make(map[core.HashT]uint64)
	for _, txId := range txIds {
		if !s.inv.HasTx(txId) {
			continue
		}
		blockId, ok := s.includedTxBlocks[txId]
		if !ok {
			out[txId] = 0
			continue
		}
		depth := s.inv.GetBlockHeight(s.head) - s.inv.GetBlockHeight(blockId)
		out[txId] = depth + 1 // Depth of 0 means it's been included once, so add 1
	}
	return out
}
