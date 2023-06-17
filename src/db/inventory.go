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
	HasBlock(blockId HashT2) bool
	HasAnyBlock(blockIds []HashT2) (HashT2, bool)
	GetBlock(blockId HashT2) Block
	GetBlockHeight(blockId HashT2) uint64
	GetBlockTotalWork(blockId HashT2) HashT2
	GetBlockParentId(blockId HashT2) HashT2
	GetBlockAncestors(blockId HashT2, maxLen int) []HashT2
	GetBlockAncestorDepth(blockId, ancestorId HashT2) (uint64, bool)
	GetBlockLCA(blockId, otherBlockId HashT2) HashT2
	HasMerkle(nodeId HashT2) bool
	GetMerkle(merkleId HashT2) MerkleNode
	GetMerkleVSize(merkleId HashT2) uint64
	GetMerkleTxIds(root HashT2) []HashT2
	GetMerkleTxs(root HashT2) []Tx
	HasTx(txId HashT2) bool
	GetTx(txId HashT2) Tx
	GetTxVSize(txId HashT2) uint64
	HasTxOut(txId HashT2, ind uint64) bool
	GetTxOut(txId HashT2, ind uint64) TxOut
	HasEntity(entityId HashT2) bool
	GetEntityVSize(entityId HashT2) uint64
}

type blockRecord struct {
	block     Block
	height    uint64
	totalWork HashT2
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
	blocks  *util.SyncMap[HashT2, blockRecord]
	merkles *util.SyncMap[HashT2, merkleRecord]
	txs     *util.SyncMap[HashT2, txRecord]
}

func NewInv() *Inv {
	inv := &Inv{
		blocks:  util.NewSyncMap[HashT2, blockRecord](),
		merkles: util.NewSyncMap[HashT2, merkleRecord](),
		txs:     util.NewSyncMap[HashT2, txRecord](),
	}
	inv.blocks.Store(HashT2{}, blockRecord{
		block:     Block{},
		height:    0,
		totalWork: HashT2{},
	})
	return inv
}

// Return whether the given block id exists.
func (inv *Inv) HasBlock(blockId HashT2) bool {
	return inv.blocks.Has(blockId)
}

func (inv *Inv) HasAnyBlock(blockIds []HashT2) (HashT2, bool) {
	for i := 0; i < len(blockIds); i++ {
		if inv.HasBlock(blockIds[i]) {
			return blockIds[i], true
		}
	}
	return HashT2{}, false
}

// Get a block, panic if it doesn't exist.
func (inv *Inv) GetBlock(blockId HashT2) Block {
	return inv.blocks.Get(blockId).block
}

// Get a block's height (0x0 is height 0, origin block is height 1).
func (inv *Inv) GetBlockHeight(blockId HashT2) uint64 {
	return inv.blocks.Get(blockId).height
}

// Get total work along chain terminating with this block.
func (inv *Inv) GetBlockTotalWork(blockId HashT2) HashT2 {
	return inv.blocks.Get(blockId).totalWork
}

func (inv *Inv) GetBlockParentId(blockId HashT2) HashT2 {
	if blockId.EqZero() {
		panic("Cannot get parent of root block")
	}
	return inv.GetBlock(blockId).PrevBlockId
}

func (inv *Inv) GetBlockAncestors(blockId HashT2, maxLen int) []HashT2 {
	out := make([]HashT2, 0)
	next := blockId
	for i := 0; i < maxLen; i++ {
		next = inv.GetBlockParentId(next)
		out = append(out, next)
		if next.EqZero() {
			break
		}
	}
	return out
}

// Returns how many blocks deep the ancestor is, and whether we have this ancestor.
func (inv *Inv) GetBlockAncestorDepth(blockId, ancestorId HashT2) (uint64, bool) {
	depth := uint64(0)
	for blockId != ancestorId && !blockId.EqZero() {
		blockId = inv.GetBlockParentId(blockId)
		depth += 1
	}
	if blockId != ancestorId {
		return 0, false
	}
	return depth, true
}

// Gets block ancestors, from top, until the given block.
// Don't include either blockId or untilId.
func (inv *Inv) GetBlockAncestorsUntil(blockId, untilId HashT2) []HashT2 {
	depth, ok := inv.GetBlockAncestorDepth(blockId, untilId)
	if !ok {
		panic("block does not have ancestor")
	}
	if depth <= 1 {
		return []HashT2{}
	}
	out := make([]HashT2, depth-1)
	for i := range out {
		if blockId.EqZero() || blockId == untilId {
			panic("exceeded expected ancestor depth")
		}
		blockId = inv.GetBlockParentId(blockId)
		out[i] = blockId
	}
	if inv.GetBlockParentId(blockId) != untilId {
		panic("last ancestor should be given untilId")
	}
	return out
}

// Return the most recent common ancestor of the two block ids.
func (inv *Inv) GetBlockLCA(blockId, otherBlockId HashT2) HashT2 {
	// Move the higher block down until it's even with the other
	for inv.GetBlockHeight(blockId) > inv.GetBlockHeight(otherBlockId) {
		blockId = inv.GetBlockParentId(blockId)
	}
	for inv.GetBlockHeight(otherBlockId) > inv.GetBlockHeight(blockId) {
		otherBlockId = inv.GetBlockParentId(otherBlockId)
	}
	// Step both blocks until they're even
	for blockId != otherBlockId {
		blockId = inv.GetBlockParentId(blockId)
		otherBlockId = inv.GetBlockParentId(otherBlockId)
	}
	return blockId
}

// Return whether the given merkle id exists.
func (inv *Inv) HasMerkle(nodeId HashT2) bool {
	return inv.merkles.Has(nodeId)
}

// Get a merkle, panic if it doesn't exist.
func (inv *Inv) GetMerkle(merkleId HashT2) MerkleNode {
	return inv.merkles.Get(merkleId).merkle
}

// Get the vSize of all txs descended from a merkle node, panic if it doesn't exist.
func (inv *Inv) GetMerkleVSize(merkleId HashT2) uint64 {
	return inv.merkles.Get(merkleId).vSize
}

// Load ids of all txs descended from a merkle node.
func (inv *Inv) GetMerkleTxIds(root HashT2) []HashT2 {
	outTxIds := make([]HashT2, 0)
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
			panic(fmt.Sprintf("unrecognized tree node: %s", nextId))
		}
	}
	return outTxIds
}

// Load all txs descended from a merkle node.
func (inv *Inv) GetMerkleTxs(root HashT2) []Tx {
	txIds := inv.GetMerkleTxIds(root)
	out := make([]Tx, len(txIds))
	for i, txId := range txIds {
		out[i] = inv.GetTx(txId)
	}
	return out
}

// Return whether the given tx id exists.
func (inv *Inv) HasTx(txId HashT2) bool {
	return inv.txs.Has(txId)
}

// Get a tx, panic if it doesn't exist.
func (inv *Inv) GetTx(txId HashT2) Tx {
	return inv.txs.Get(txId).tx
}

// Get a tx's vSize, panic if it doesn't exist.
func (inv *Inv) GetTxVSize(txId HashT2) uint64 {
	return inv.txs.Get(txId).vSize
}

// Return whether the given tx has the given output index.
func (inv *Inv) HasTxOut(txId HashT2, ind uint64) bool {
	if !inv.HasTx(txId) {
		return false
	}
	return ind < uint64(len(inv.GetTx(txId).Outputs))
}

// Get the given output from the given tx.
func (inv *Inv) GetTxOut(txId HashT2, ind uint64) TxOut {
	return inv.GetTx(txId).Outputs[ind]
}

// Return whether the given id exists as either a merkle or a tx.
func (inv *Inv) HasEntity(entityId HashT2) bool {
	return inv.HasMerkle(entityId) || inv.HasTx(entityId)
}

// Return the vSize of the given merkle or tx, panic if neither exists.
func (inv *Inv) GetEntityVSize(entityId HashT2) uint64 {
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
		return fmt.Errorf("new block already known: %s", blockId)
	}
	if err := verifyBlock(inv, block); err != nil {
		return err
	}
	prevWork := inv.GetBlockTotalWork(block.PrevBlockId)
	inv.blocks.Store(blockId, blockRecord{
		block:     block,
		height:    inv.GetBlockHeight(block.PrevBlockId) + 1,
		totalWork: prevWork.WorkAppendTarget(block.Difficulty),
	})
	return nil
}

// Verify and store a new merkle node.
func (inv *Inv) StoreMerkle(merkle MerkleNode) error {
	nodeId := merkle.Hash()
	if inv.HasMerkle(nodeId) {
		return fmt.Errorf("merkle already known: %s", nodeId)
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
		return fmt.Errorf("tx already known: %s", txId)
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
