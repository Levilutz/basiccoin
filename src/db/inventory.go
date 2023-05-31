package db

import (
	"errors"
	"fmt"

	"github.com/levilutz/basiccoin/src/util"
)

var ErrEntityKnown = errors.New("entity known")
var ErrEntityUnknown = errors.New("entity unknown")

// Interface of all the functions that can't invoke SyncMap.Store.
type InvReader interface {
	HasBlock(blockId HashT) bool
	HasAnyBlock(blockIds []HashT) (HashT, bool)
	GetBlock(blockId HashT) Block
	GetBlockHeight(blockId HashT) uint64
	GetBlockParentId(blockId HashT) HashT
	GetBlockAncestors(blockId HashT, maxLen int) []HashT
	GetBlockAncestorDepth(blockId HashT, ancestorId HashT) (uint64, bool)
	HasMerkle(nodeId HashT) bool
	GetMerkle(merkleId HashT) MerkleNode
	GetMerkleTxs(root HashT) []Tx
	HasTx(txId HashT) bool
	GetTx(txId HashT) Tx
	HasTxOut(txId HashT, ind uint64) bool
	GetTxOut(txId HashT, ind uint64) TxOut
}

// Write-once read-many maps.
// Only one thread should be making writes at a time, but many can be reading.
type Inv struct {
	// Main inventory
	blocks  *util.SyncMap[HashT, Block]
	merkles *util.SyncMap[HashT, MerkleNode]
	txs     *util.SyncMap[HashT, Tx]
	// Aux info (must be inserted before referenced main entity)
	blockHs *util.SyncMap[HashT, uint64]
}

func NewInv() *Inv {
	inv := &Inv{
		blocks:  util.NewSyncMap[HashT, Block](),
		merkles: util.NewSyncMap[HashT, MerkleNode](),
		txs:     util.NewSyncMap[HashT, Tx](),
		blockHs: util.NewSyncMap[HashT, uint64](),
	}
	inv.blockHs.Store(HashTZero, 0)
	inv.blocks.Store(HashTZero, Block{})
	return inv
}

// Return whether the given block id exists.
func (inv *Inv) HasBlock(blockId HashT) bool {
	_, ok := inv.blocks.Load(blockId)
	return ok
}

func (inv *Inv) HasAnyBlock(blockIds []HashT) (HashT, bool) {
	for i := 0; i < len(blockIds); i++ {
		if inv.HasBlock(blockIds[i]) {
			return blockIds[i], true
		}
	}
	return HashTZero, false
}

// Get a block, panic if it doesn't exist.
func (inv *Inv) GetBlock(blockId HashT) Block {
	block, ok := inv.blocks.Load(blockId)
	if !ok {
		panic(fmt.Sprintf("block should exist: %x", blockId))
	}
	return block
}

// Get a block's height (0x0 is height 0, origin block is height 1).
func (inv *Inv) GetBlockHeight(blockId HashT) uint64 {
	h, ok := inv.blockHs.Load(blockId)
	if !ok {
		panic(fmt.Sprintf("block should exist: %x", blockId))
	}
	return h
}

func (inv *Inv) GetBlockParentId(blockId HashT) HashT {
	if blockId == HashTZero {
		panic("Cannot get parent of root block")
	}
	return inv.GetBlock(blockId).PrevBlockId
}

func (inv *Inv) GetBlockAncestors(blockId HashT, maxLen int) []HashT {
	out := make([]HashT, 0)
	next := blockId
	for i := 0; i < maxLen; i++ {
		next = inv.GetBlockParentId(next)
		out = append(out, next)
		if next == HashTZero {
			break
		}
	}
	return out
}

// Returns how many blocks deep the ancestor is, and whether we have this ancestor.
func (inv *Inv) GetBlockAncestorDepth(blockId, ancestorId HashT) (uint64, bool) {
	depth := uint64(0)
	for blockId != ancestorId && blockId != HashTZero {
		blockId = inv.GetBlockParentId(blockId)
	}
	if blockId != ancestorId {
		return 0, false
	}
	return depth, true
}

// Return whether the given merkle id exists.
func (inv *Inv) HasMerkle(nodeId HashT) bool {
	_, ok := inv.merkles.Load(nodeId)
	return ok
}

// Get a merkle, panic if it doesn't exist.
func (inv *Inv) GetMerkle(merkleId HashT) MerkleNode {
	merkle, ok := inv.merkles.Load(merkleId)
	if !ok {
		panic(fmt.Sprintf("merkle should exist: %x", merkleId))
	}
	return merkle
}

// Load all txs descended from a merkle node.
func (inv *Inv) GetMerkleTxs(root HashT) []Tx {
	outTxs := make([]Tx, 0)
	// Go through each node in tree, categorizing as either tx or merkle
	idQueue := util.NewQueue[HashT]()
	visitedIds := util.NewSet[HashT]() // Prevent cycles
	idQueue.Push(root)
	for i := 0; i < int(MerkleTreeMaxSize()); i++ {
		// Pop next id, finish if we've cleared queue
		nextId, ok := idQueue.Pop()
		if !ok {
			break
		}

		// Prevent cycles
		if visitedIds.Includes(nextId) {
			panic(fmt.Sprintf("circular visit to id %x", nextId))
		}
		visitedIds.Add(nextId)

		// Load tx or merkle and categorize
		if inv.HasTx(nextId) {
			outTxs = append(outTxs, inv.GetTx(nextId))
		} else if inv.HasMerkle(nextId) {
			merkle := inv.GetMerkle(nextId)
			idQueue.Push(merkle.LChild)
			if merkle.RChild != merkle.LChild {
				idQueue.Push(merkle.RChild)
			}
		} else {
			// TODO requires verification on merkle insert that subtrees don't break limit
			panic(fmt.Sprintf("unrecognized tree node: %x", nextId))
		}
	}

	// Verify we didn't just hit limit
	if idQueue.Size() > 0 {
		panic(fmt.Sprintf("tree exceeds max size of %d", MerkleTreeMaxSize()))
	}
	return outTxs
}

// Return whether the given tx id exists.
func (inv *Inv) HasTx(txId HashT) bool {
	_, ok := inv.txs.Load(txId)
	return ok
}

// Get a tx, panic if it doesn't exist.
func (inv *Inv) GetTx(txId HashT) Tx {
	tx, ok := inv.txs.Load(txId)
	if !ok {
		panic(fmt.Sprintf("tx should exist: %x", txId))
	}
	return tx
}

// Return whether the given tx has the given output index.
func (inv *Inv) HasTxOut(txId HashT, ind uint64) bool {
	if !inv.HasTx(txId) {
		return false
	}
	return ind >= uint64(len(inv.GetTx(txId).Outputs))
}

// Get the given output from the given tx.
func (inv *Inv) GetTxOut(txId HashT, ind uint64) TxOut {
	return inv.GetTx(txId).Outputs[ind]
}

// Store a new block, ensures merkle root known and difficulty beat.
func (inv *Inv) StoreBlock(b Block) error {
	// Pre-insertion verification
	blockId := b.Hash()
	if inv.HasBlock(blockId) {
		return fmt.Errorf("block already known: %x", blockId)
	} else if !BelowTarget(blockId, b.Difficulty) {
		return fmt.Errorf("block failed to beat target difficulty")
	} else if !inv.HasMerkle(b.MerkleRoot) {
		return fmt.Errorf("failed to find new block merkle root")
	}
	parentHeight := inv.GetBlockHeight(b.PrevBlockId)
	txs := inv.GetMerkleTxs(b.MerkleRoot)
	if len(txs) == 0 {
		return fmt.Errorf("block has no txs")
	} else if len(txs) > int(BlockMaxTxs()) {
		return fmt.Errorf("block has too many txs")
	} else if len(txs[0].Inputs) != 0 {
		return fmt.Errorf("block missing coinbase tx")
	} else if txs[0].MinBlock != parentHeight+1 {
		return fmt.Errorf("coinbase MinBlock does not equal height")
	}
	totalInputs := uint64(util.Constants.BlockReward)
	totalOutputs := uint64(0)
	totalVSize := uint64(0)
	for i, tx := range txs {
		if i == 0 {
			totalOutputs += tx.TotalOutputs()
		} else {
			totalInputs += tx.TotalInputs()
			totalOutputs += tx.TotalOutputs()
		}
		totalVSize += tx.VSize()
	}
	if totalInputs != totalOutputs {
		return fmt.Errorf("total inputs and outputs do not match")
	} else if totalVSize > util.Constants.MaxBlockVSize {
		return fmt.Errorf("block exceeds max vSize")
	}
	// Insert
	inv.blockHs.Store(blockId, parentHeight+1)
	inv.blocks.Store(blockId, b)
	return nil
}

// Store a new merkle, ensures children known.
func (inv *Inv) StoreMerkle(merkle MerkleNode) error {
	// Pre-insertion verification
	nodeId := merkle.Hash()
	if inv.HasMerkle(nodeId) {
		return fmt.Errorf("merkle already known: %x", nodeId)
	} else if !inv.HasMerkle(merkle.LChild) && !inv.HasTx(merkle.LChild) {
		return fmt.Errorf("failed to find LChild: %x", merkle.LChild)
	} else if !inv.HasMerkle(merkle.RChild) && !inv.HasTx(merkle.RChild) {
		return fmt.Errorf("failed to find RChild: %x", merkle.RChild)
	}
	// Insert
	inv.merkles.Store(nodeId, merkle)
	return nil
}

// Store a new transaction.
func (inv *Inv) StoreTx(tx Tx) error {
	// Pre-insertion verification
	txId := tx.Hash()
	if inv.HasTx(txId) {
		return fmt.Errorf("tx already known: %x", txId)
	}
	if !tx.SignaturesValid() {
		return fmt.Errorf("tx signatures invalid")
	} else if tx.VSize() > util.Constants.MaxTxVSize {
		return fmt.Errorf("tx VSize exceeds limit")
	}
	if len(tx.Inputs) > 0 {
		// Not coinbase - verify total outputs < total inputs
		if tx.TotalOutputs() >= tx.TotalInputs() {
			return fmt.Errorf("tx outputs exceed or match inputs")
		}
	} else {
		// Coinbase - verify outputs exist and total outputs >= BlockReward
		if len(tx.Outputs) != 1 {
			return fmt.Errorf("coinbase must have 1 output")
		} else if tx.TotalOutputs() < uint64(util.Constants.BlockReward) {
			return fmt.Errorf("coinbase has insufficient block reward")
		}
	}
	// Verify given public key and value match those on origin utxo
	for _, txi := range tx.Inputs {
		if !inv.HasTxOut(txi.OriginTxId, txi.OriginTxOutInd) {
			return fmt.Errorf(
				"failed to find utxo %x[%d]",
				txi.OriginTxId,
				txi.OriginTxOutInd,
			)
		}
		origin := inv.GetTxOut(txi.OriginTxId, txi.OriginTxOutInd)
		if DHash(txi.PublicKey) != origin.PublicKeyHash {
			return fmt.Errorf("given public key does not match claimed utxo")
		}
		if txi.Value != origin.Value {
			return fmt.Errorf("given value does not match claimed utxo")
		}
	}
	// Insert
	inv.txs.Store(txId, tx)
	return nil
}
