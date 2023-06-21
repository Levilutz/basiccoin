package inv

import (
	"fmt"

	"github.com/levilutz/basiccoin/pkg/core"
	"github.com/levilutz/basiccoin/pkg/queue"
	"github.com/levilutz/basiccoin/pkg/syncmap"
)

// Interface of all the inv methods that can't invoke SyncMap.Store.
type InvReader interface {
	GetBlock(blockId core.HashT) core.Block
	GetBlockAncestorDepth(blockId core.HashT, ancestorId core.HashT) (uint64, bool)
	GetBlockAncestors(blockId core.HashT, maxLen int) []core.HashT
	GetBlockAncestorsUntil(blockId core.HashT, untilId core.HashT) []core.HashT
	GetBlockHeight(blockId core.HashT) uint64
	GetBlockLCA(blockId core.HashT, otherBlockId core.HashT) core.HashT
	GetBlockParentId(blockId core.HashT) core.HashT
	GetBlockSpecificAncestor(blockId core.HashT, depth int) core.HashT
	GetBlockTotalWork(blockId core.HashT) core.HashT
	GetCoreParams() core.Params
	GetEntityVSize(entityId core.HashT) uint64
	GetMerkle(merkleId core.HashT) core.MerkleNode
	GetMerkleTxIds(root core.HashT) []core.HashT
	GetMerkleTxs(root core.HashT) []core.Tx
	GetMerkleVSize(merkleId core.HashT) uint64
	GetTx(txId core.HashT) core.Tx
	GetTxOut(txId core.HashT, ind uint64) core.TxOut
	GetTxVSize(txId core.HashT) uint64
	HasAnyBlock(blockIds []core.HashT) (core.HashT, bool)
	HasBlock(blockId core.HashT) bool
	HasEntity(entityId core.HashT) bool
	HasMerkle(nodeId core.HashT) bool
	HasTx(txId core.HashT) bool
	HasTxOut(txId core.HashT, ind uint64) bool
}

type blockRecord struct {
	block     core.Block
	height    uint64
	totalWork core.HashT
}

type merkleRecord struct {
	merkle core.MerkleNode
	vSize  uint64
}

type txRecord struct {
	tx    core.Tx
	vSize uint64
}

// A shared inventory of various entities.
// Only one goroutine should be allowed to make writes at a time.
type Inv struct {
	coreParams core.Params
	verifier   *core.Verifier
	// Main inventory
	blocks  *syncmap.SyncMap[core.HashT, blockRecord]
	merkles *syncmap.SyncMap[core.HashT, merkleRecord]
	txs     *syncmap.SyncMap[core.HashT, txRecord]
}

func NewInv(coreParams core.Params) *Inv {
	inv := &Inv{
		coreParams: coreParams,
		blocks:     syncmap.NewSyncMap[core.HashT, blockRecord](),
		merkles:    syncmap.NewSyncMap[core.HashT, merkleRecord](),
		txs:        syncmap.NewSyncMap[core.HashT, txRecord](),
	}
	inv.verifier = core.NewVerifier(coreParams, inv)
	inv.blocks.Store(core.HashT{}, blockRecord{
		block:     core.Block{},
		height:    0,
		totalWork: core.HashT{},
	})
	return inv
}

// Get the stored core params.
func (inv *Inv) GetCoreParams() core.Params {
	return inv.coreParams
}

// Return whether the given block id exists.
func (inv *Inv) HasBlock(blockId core.HashT) bool {
	return inv.blocks.Has(blockId)
}

func (inv *Inv) HasAnyBlock(blockIds []core.HashT) (core.HashT, bool) {
	for i := 0; i < len(blockIds); i++ {
		if inv.HasBlock(blockIds[i]) {
			return blockIds[i], true
		}
	}
	return core.HashT{}, false
}

// Get a block, panic if it doesn't exist.
func (inv *Inv) GetBlock(blockId core.HashT) core.Block {
	return inv.blocks.Get(blockId).block
}

// Get a block's height (0x0 is height 0, origin block is height 1).
func (inv *Inv) GetBlockHeight(blockId core.HashT) uint64 {
	return inv.blocks.Get(blockId).height
}

// Get total work along chain terminating with this block.
func (inv *Inv) GetBlockTotalWork(blockId core.HashT) core.HashT {
	return inv.blocks.Get(blockId).totalWork
}

func (inv *Inv) GetBlockParentId(blockId core.HashT) core.HashT {
	if blockId.EqZero() {
		panic("Cannot get parent of root block")
	}
	return inv.GetBlock(blockId).PrevBlockId
}

// Get up to maxLen of this block's ancestors. Does not include the given block.
// Includes the zero block if within maxLen ancestors.
func (inv *Inv) GetBlockAncestors(blockId core.HashT, maxLen int) []core.HashT {
	if blockId.EqZero() {
		panic("Cannot get ancestors of root block")
	}
	out := make([]core.HashT, 0)
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

// Get the blockId that's `depth` hops up from here, or the zero block we hit chain start first.
func (inv *Inv) GetBlockSpecificAncestor(blockId core.HashT, depth int) core.HashT {
	// If no hops or we're already at zero, return this block
	if depth == 0 || blockId.EqZero() {
		return blockId
	}
	ancestorIds := inv.GetBlockAncestors(blockId, depth)
	// Since blockId != 0, we know we have at least one ancestor
	return ancestorIds[len(ancestorIds)-1]
}

// Returns how many blocks deep the ancestor is, and whether we have this ancestor.
func (inv *Inv) GetBlockAncestorDepth(blockId, ancestorId core.HashT) (uint64, bool) {
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
func (inv *Inv) GetBlockAncestorsUntil(blockId, untilId core.HashT) []core.HashT {
	depth, ok := inv.GetBlockAncestorDepth(blockId, untilId)
	if !ok {
		panic("block does not have ancestor")
	}
	if depth <= 1 {
		return []core.HashT{}
	}
	out := make([]core.HashT, depth-1)
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
func (inv *Inv) GetBlockLCA(blockId, otherBlockId core.HashT) core.HashT {
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
func (inv *Inv) HasMerkle(nodeId core.HashT) bool {
	return inv.merkles.Has(nodeId)
}

// Get a merkle, panic if it doesn't exist.
func (inv *Inv) GetMerkle(merkleId core.HashT) core.MerkleNode {
	return inv.merkles.Get(merkleId).merkle
}

// Get the vSize of all txs descended from a merkle node, panic if it doesn't exist.
func (inv *Inv) GetMerkleVSize(merkleId core.HashT) uint64 {
	return inv.merkles.Get(merkleId).vSize
}

// Load ids of all txs descended from a merkle node.
func (inv *Inv) GetMerkleTxIds(root core.HashT) []core.HashT {
	outTxIds := make([]core.HashT, 0)
	// Go through each node in tree, categorizing as either tx or merkle
	idQueue := queue.NewQueue(root)
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
func (inv *Inv) GetMerkleTxs(root core.HashT) []core.Tx {
	txIds := inv.GetMerkleTxIds(root)
	out := make([]core.Tx, len(txIds))
	for i, txId := range txIds {
		out[i] = inv.GetTx(txId)
	}
	return out
}

// Return whether the given tx id exists.
func (inv *Inv) HasTx(txId core.HashT) bool {
	return inv.txs.Has(txId)
}

// Get a tx, panic if it doesn't exist.
func (inv *Inv) GetTx(txId core.HashT) core.Tx {
	return inv.txs.Get(txId).tx
}

// Get a tx's vSize, panic if it doesn't exist.
func (inv *Inv) GetTxVSize(txId core.HashT) uint64 {
	return inv.txs.Get(txId).vSize
}

// Return whether the given tx has the given output index.
func (inv *Inv) HasTxOut(txId core.HashT, ind uint64) bool {
	if !inv.HasTx(txId) {
		return false
	}
	return ind < uint64(len(inv.GetTx(txId).Outputs))
}

// Get the given output from the given tx.
func (inv *Inv) GetTxOut(txId core.HashT, ind uint64) core.TxOut {
	return inv.GetTx(txId).Outputs[ind]
}

// Return whether the given id exists as either a merkle or a tx.
func (inv *Inv) HasEntity(entityId core.HashT) bool {
	return inv.HasMerkle(entityId) || inv.HasTx(entityId)
}

// Return the vSize of the given merkle or tx, panic if neither exists.
func (inv *Inv) GetEntityVSize(entityId core.HashT) uint64 {
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
func (inv *Inv) StoreBlock(block core.Block) error {
	blockId := block.Hash()
	if inv.HasBlock(blockId) {
		return fmt.Errorf("new block already known: %s", blockId)
	}
	if err := inv.verifier.VerifyBlock(block); err != nil {
		return err
	}
	prevWork := inv.GetBlockTotalWork(block.PrevBlockId)
	inv.blocks.Store(blockId, blockRecord{
		block:     block,
		height:    inv.GetBlockHeight(block.PrevBlockId) + 1,
		totalWork: prevWork.WorkAppendTarget(block.Target),
	})
	return nil
}

// Verify and store a new merkle node.
func (inv *Inv) StoreMerkle(merkle core.MerkleNode) error {
	nodeId := merkle.Hash()
	if inv.HasMerkle(nodeId) {
		return fmt.Errorf("merkle already known: %s", nodeId)
	}
	if err := inv.verifier.VerifyMerkle(merkle); err != nil {
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
func (inv *Inv) StoreTx(tx core.Tx) error {
	txId := tx.Hash()
	if inv.HasTx(txId) {
		return fmt.Errorf("tx already known: %s", txId)
	}
	if err := inv.verifier.VerifyTx(tx); err != nil {
		return err
	}
	inv.txs.Store(txId, txRecord{
		tx:    tx,
		vSize: tx.VSize(),
	})
	return nil
}
