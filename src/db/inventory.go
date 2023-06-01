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
	GetBlockTotalWork(blockId HashT) HashT
	GetBlockParentId(blockId HashT) HashT
	GetBlockAncestors(blockId HashT, maxLen int) []HashT
	GetBlockAncestorDepth(blockId HashT, ancestorId HashT) (uint64, bool)
	HasMerkle(nodeId HashT) bool
	GetMerkle(merkleId HashT) MerkleNode
	GetMerkleVSize(merkleId HashT) uint64
	GetMerkleTxIds(root HashT) []HashT
	GetMerkleTxs(root HashT) []Tx
	HasTx(txId HashT) bool
	GetTx(txId HashT) Tx
	GetTxVSize(txId HashT) uint64
	HasTxOut(txId HashT, ind uint64) bool
	GetTxOut(txId HashT, ind uint64) TxOut
	HasEntity(entityId HashT) bool
	GetEntityVSize(entityId HashT) uint64
}

type blockRecord struct {
	block     Block
	height    uint64
	totalWork HashT
}

type merkleRecord struct {
	merkle MerkleNode
	vSize  uint64
}

type txRecord struct {
	tx    Tx
	vSize uint64
}

// Write-once read-many maps.
// Only one thread should be making writes at a time, but many can be reading.
type Inv struct {
	// Main inventory
	blocks  *util.SyncMap[HashT, blockRecord]
	merkles *util.SyncMap[HashT, merkleRecord]
	txs     *util.SyncMap[HashT, txRecord]
}

func NewInv() *Inv {
	inv := &Inv{
		blocks:  util.NewSyncMap[HashT, blockRecord](),
		merkles: util.NewSyncMap[HashT, merkleRecord](),
		txs:     util.NewSyncMap[HashT, txRecord](),
	}
	inv.blocks.Store(HashTZero, blockRecord{
		block:     Block{},
		height:    0,
		totalWork: HashTZero,
	})
	return inv
}

// Return whether the given block id exists.
func (inv *Inv) HasBlock(blockId HashT) bool {
	return inv.blocks.Has(blockId)
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
	return inv.blocks.Get(blockId).block
}

// Get a block's height (0x0 is height 0, origin block is height 1).
func (inv *Inv) GetBlockHeight(blockId HashT) uint64 {
	return inv.blocks.Get(blockId).height
}

// Get total work along chain terminating with this block.
func (inv *Inv) GetBlockTotalWork(blockId HashT) HashT {
	return inv.blocks.Get(blockId).totalWork
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
	return inv.merkles.Has(nodeId)
}

// Get a merkle, panic if it doesn't exist.
func (inv *Inv) GetMerkle(merkleId HashT) MerkleNode {
	return inv.merkles.Get(merkleId).merkle
}

// Get the vSize of all txs descended from a merkle node, panic if it doesn't exist.
func (inv *Inv) GetMerkleVSize(merkleId HashT) uint64 {
	return inv.merkles.Get(merkleId).vSize
}

// Load ids of all txs descended from a merkle node.
func (inv *Inv) GetMerkleTxIds(root HashT) []HashT {
	outTxIds := make([]HashT, 0)
	// Go through each node in tree, categorizing as either tx or merkle
	idQueue := util.NewQueue(root)
	// Pick off queue until empty (finite bc merkle tree can't be cyclic)
	for idQueue.Size() > 0 {
		nextId, _ := idQueue.Pop()
		// Load tx or merkle and categorize
		if inv.HasTx(nextId) {
			outTxIds = append(outTxIds, nextId)
		} else if inv.HasMerkle(nextId) {
			merkle := inv.GetMerkle(nextId)
			idQueue.Push(merkle.LChild)
			if merkle.RChild != merkle.LChild {
				idQueue.Push(merkle.RChild)
			}
		} else {
			panic(fmt.Sprintf("unrecognized tree node: %x", nextId))
		}
	}
	return outTxIds
}

// Load all txs descended from a merkle node.
func (inv *Inv) GetMerkleTxs(root HashT) []Tx {
	txIds := inv.GetMerkleTxIds(root)
	out := make([]Tx, len(txIds))
	for i, txId := range txIds {
		out[i] = inv.GetTx(txId)
	}
	return out
}

// Return whether the given tx id exists.
func (inv *Inv) HasTx(txId HashT) bool {
	return inv.txs.Has(txId)
}

// Get a tx, panic if it doesn't exist.
func (inv *Inv) GetTx(txId HashT) Tx {
	return inv.txs.Get(txId).tx
}

// Get a tx's vSize, panic if it doesn't exist.
func (inv *Inv) GetTxVSize(txId HashT) uint64 {
	return inv.txs.Get(txId).vSize
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

// Return whether the given id exists as either a merkle or a tx.
func (inv *Inv) HasEntity(entityId HashT) bool {
	return inv.HasMerkle(entityId) || inv.HasTx(entityId)
}

// Return the vSize of the given merkle or tx, panic if neither exists.
func (inv *Inv) GetEntityVSize(entityId HashT) uint64 {
	if inv.HasMerkle(entityId) {
		return inv.GetMerkleVSize(entityId)
	}
	return inv.GetTxVSize(entityId)
}

// Verify and store a new block.
// For efficiency, this won't verify that each tx's claimed utxos are available.
// Thus the caller (usually a State) should verify to prevent double-spends.
// This verification usually wouldn't be done at block-insertion time
// (as utxos only exist in the context of a full chain),
// but rather should be done when attempting to advance the chain head.
func (inv *Inv) StoreBlock(block Block) error {
	blockId := block.Hash()
	if inv.HasBlock(blockId) {
		return fmt.Errorf("new block already known: %x", blockId)
	}
	if err := verifyBlock(inv, block); err != nil {
		return err
	}
	prevWork := inv.GetBlockTotalWork(block.PrevBlockId)
	inv.blocks.Store(blockId, blockRecord{
		block:     block,
		height:    inv.GetBlockHeight(block.PrevBlockId) + 1,
		totalWork: AppendTotalWork(prevWork, block.Difficulty),
	})
	return nil
}

// Verify and store a new merkle node.
func (inv *Inv) StoreMerkle(merkle MerkleNode) error {
	nodeId := merkle.Hash()
	if inv.HasMerkle(nodeId) {
		return fmt.Errorf("merkle already known: %x", nodeId)
	}
	if err := verifyMerkle(inv, merkle); err != nil {
		return err
	}
	totalSize := inv.GetEntityVSize(merkle.LChild)
	if merkle.RChild != merkle.LChild {
		totalSize += inv.GetEntityVSize(merkle.RChild)
	}
	inv.merkles.Store(nodeId, merkleRecord{
		merkle: merkle,
		vSize:  totalSize,
	})
	return nil
}

// Verify and store a new transaction.
func (inv *Inv) StoreTx(tx Tx) error {
	txId := tx.Hash()
	if inv.HasTx(txId) {
		return fmt.Errorf("tx already known: %x", txId)
	}
	if err := verifyTx(inv, tx); err != nil {
		return err
	}
	inv.txs.Store(txId, txRecord{
		tx:    tx,
		vSize: tx.VSize(),
	})
	return nil
}
