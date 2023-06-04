package db

import (
	"fmt"
	"sort"

	"github.com/levilutz/basiccoin/src/util"
)

// A record of balance originating from a utxo.
type BalanceRecord struct {
	TxId  HashT
	Ind   uint64
	Value uint64
}

// State at a blockchain node. Responsible for preventing double-spends.
// Meant to only be accessed synchronously by a single thread.
type State struct {
	head         HashT
	mempool      *util.Set[HashT]
	utxos        *util.Set[Utxo]
	inv          InvReader
	mempoolRates map[HashT]float64
	balances     map[HashT][]BalanceRecord
}

func NewState(inv InvReader, trackBalances bool) *State {
	s := &State{
		head:         HashTZero,
		mempool:      util.NewSet[HashT](),
		utxos:        util.NewSet[Utxo](),
		inv:          inv,
		mempoolRates: make(map[HashT]float64),
		balances:     nil,
	}
	if trackBalances {
		s.balances = make(map[HashT][]BalanceRecord)
	}
	return s
}

// Copy a state.
func (s *State) Copy() *State {
	return &State{
		head:         s.head,
		mempool:      s.mempool.Copy(),
		utxos:        s.utxos.Copy(),
		inv:          s.inv,
		mempoolRates: util.CopyMap(s.mempoolRates),
		balances:     util.CopyMap(s.balances),
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
		for _, utxo := range tx.GetConsumedUtxos() {
			s.utxos.Add(utxo)
			if s.balances != nil {
				txo := s.inv.GetTxOut(utxo.TxId, utxo.Ind)
				s.creditBalance(txo.PublicKeyHash, BalanceRecord{
					TxId:  utxo.TxId,
					Ind:   utxo.Ind,
					Value: txo.Value,
				})
			}
		}
		// Remove the tx outputs from the utxo set
		for i, txo := range tx.Outputs {
			if !s.utxos.Remove(Utxo{TxId: txId, Ind: uint64(i)}) {
				return fmt.Errorf("state corrupt - missing utxo %x[%d]", txId, i)
			}
			if s.balances != nil {
				s.debitBalance(txo.PublicKeyHash, BalanceRecord{
					TxId:  txId,
					Ind:   uint64(i),
					Value: txo.Value,
				})
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
		for _, utxo := range tx.GetConsumedUtxos() {
			if !s.utxos.Remove(utxo) {
				return fmt.Errorf("tx input not available %x[%d]", utxo.TxId, utxo.Ind)
			}
			if s.balances != nil {
				txo := s.inv.GetTxOut(utxo.TxId, utxo.Ind)
				s.debitBalance(txo.PublicKeyHash, BalanceRecord{
					TxId:  utxo.TxId,
					Ind:   utxo.Ind,
					Value: txo.Value,
				})
			}
		}
		// Add the tx outputs
		for i, txo := range tx.Outputs {
			s.utxos.Add(Utxo{TxId: txId, Ind: uint64(i)})
			if s.balances != nil {
				s.creditBalance(txo.PublicKeyHash, BalanceRecord{
					TxId:  txId,
					Ind:   uint64(i),
					Value: txo.Value,
				})
			}
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
	for _, utxo := range tx.GetConsumedUtxos() {
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

// Add to the balance of a public key hash.
func (s *State) creditBalance(publicKeyHash HashT, credit BalanceRecord) {
	if s.balances == nil {
		panic("balance tracking was not enabled")
	}
	_, ok := s.balances[publicKeyHash]
	if !ok {
		s.balances[publicKeyHash] = []BalanceRecord{credit}
	} else {
		s.balances[publicKeyHash] = append(s.balances[publicKeyHash], credit)
	}
}

// Remove from the balance of a public key hash.
func (s *State) debitBalance(publicKeyHash HashT, debit BalanceRecord) {
	if s.balances == nil {
		panic("balance tracking was not enabled")
	}
	balances, ok := s.balances[publicKeyHash]
	if ok {
		for i := 0; i < len(balances); i++ {
			if balances[i] == debit {
				if len(balances) == 1 {
					delete(s.balances, publicKeyHash)
				} else {
					s.balances[publicKeyHash] = append(balances[:i], balances[i+1:]...)
				}
				return
			}
		}
	}
	panic("cannot debit balance, balance does not exist")
}

// Get the balances of a public key hash.
func (s *State) GetBalances(publicKeyHash HashT) []BalanceRecord {
	if s.balances == nil {
		panic("balance tracking was not enabled")
	}
	balances, ok := s.balances[publicKeyHash]
	if !ok {
		return []BalanceRecord{}
	}
	return balances
}

func (s *State) GetTotalBalance(publicKeyHash HashT) uint64 {
	if s.balances == nil {
		panic("balance tracking was not enabled")
	}
	balances, ok := s.balances[publicKeyHash]
	if !ok {
		return 0
	}
	total := uint64(0)
	for _, balance := range balances {
		total += balance.Value
	}
	return total
}
