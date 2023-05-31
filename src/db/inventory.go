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
	GetMerkleVSize(merkleId HashT) uint64
	GetMerkleTxs(root HashT) []Tx
	HasTx(txId HashT) bool
	GetTx(txId HashT) Tx
	GetTxVSize(txId HashT) uint64
	HasTxOut(txId HashT, ind uint64) bool
	GetTxOut(txId HashT, ind uint64) TxOut
}

type blockRecord struct {
	block  Block
	height uint64
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
		block:  Block{},
		height: 0,
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

// Load all txs descended from a merkle node.
func (inv *Inv) GetMerkleTxs(root HashT) []Tx {
	outTxs := make([]Tx, 0)
	// Go through each node in tree, categorizing as either tx or merkle
	idQueue := util.NewQueue(root)
	// Pick off queue until empty (finite bc merkle tree can't be cyclic)
	for idQueue.Size() > 0 {
		nextId, _ := idQueue.Pop()
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
			panic(fmt.Sprintf("unrecognized tree node: %x", nextId))
		}
	}
	return outTxs
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

// Verify and store a new block.
// For efficiency, this won't verify that each tx's claimed utxos are available.
// Thus the caller (usually a State) should verify to prevent double-spends.
func (inv *Inv) StoreBlock(b Block) error {
	blockId := b.Hash()
	if inv.HasBlock(blockId) {
		return fmt.Errorf("new block already known: %x", blockId)
	} else if !BelowTarget(blockId, b.Difficulty) {
		return fmt.Errorf("new block failed to beat target difficulty")
	} else if !inv.HasMerkle(b.MerkleRoot) {
		return fmt.Errorf("failed to find new block merkle root")
	} else if !inv.HasBlock(b.PrevBlockId) {
		return fmt.Errorf("failed to find new block parent id")
	}
	parentHeight := inv.GetBlockHeight(b.PrevBlockId)
	txs := inv.GetMerkleTxs(b.MerkleRoot)
	if len(txs) == 0 {
		return fmt.Errorf("new block has no txs")
	} else if len(txs) > int(BlockMaxTxs()) {
		return fmt.Errorf("new block has too many txs")
	} else if txs[0].MinBlock != parentHeight+1 {
		return fmt.Errorf("coinbase MinBlock does not equal height")
	}
	totalInputs := uint64(util.Constants.BlockReward)
	totalOutputs := uint64(0)
	for i, tx := range txs {
		inputValue := tx.InputsValue()
		outputValue := tx.OutputsValue()
		if i == 0 {
			if len(tx.Inputs) != 0 || inputValue != 0 {
				return fmt.Errorf("coinbase tx must have no inputs")
			} else if len(tx.Outputs) != 1 || outputValue < util.Constants.BlockReward {
				return fmt.Errorf("coinbase tx must have outputs > block reward")
			}
		} else {
			if len(tx.Inputs) == 0 || inputValue <= outputValue {
				return fmt.Errorf("non-coinbase tx must have inputs > outputs")
			}
		}
		totalInputs += inputValue
		totalOutputs += outputValue
	}
	if totalInputs != totalOutputs {
		return fmt.Errorf("total inputs and outputs do not match")
	} else if inv.GetMerkleVSize(b.MerkleRoot) > util.Constants.MaxBlockVSize {
		return fmt.Errorf("block exceeds max vSize")
	}
	inv.blocks.Store(blockId, blockRecord{
		block:  b,
		height: parentHeight + 1,
	})
	return nil
}

// Verify and store a new merkle node.
func (inv *Inv) StoreMerkle(merkle MerkleNode) error {
	nodeId := merkle.Hash()
	if inv.HasMerkle(nodeId) {
		return fmt.Errorf("merkle already known: %x", nodeId)
	}
	// Get left child size
	totalSize := uint64(0)
	if inv.HasMerkle(merkle.LChild) {
		totalSize += inv.GetMerkleVSize(merkle.LChild)
	} else if inv.HasTx(merkle.LChild) {
		totalSize += inv.GetTxVSize(merkle.LChild)
	} else {
		return fmt.Errorf("failed to find LChild: %x", merkle.LChild)
	}
	// Get right child size (if appropriate)
	if merkle.RChild != merkle.LChild {
		if inv.HasMerkle(merkle.RChild) {
			totalSize += inv.GetMerkleVSize(merkle.RChild)
		} else if inv.HasTx(merkle.RChild) {
			totalSize += inv.GetTxVSize(merkle.RChild)
		} else {
			return fmt.Errorf("failed to find RChild: %x", merkle.RChild)
		}
	}
	if totalSize > util.Constants.MaxBlockVSize {
		return fmt.Errorf("merkle cannot be created - would exceed max block vSize")
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
	vSize := tx.VSize()
	if inv.HasTx(txId) {
		return fmt.Errorf("tx already known: %x", txId)
	} else if !tx.SignaturesValid() {
		return fmt.Errorf("tx signatures invalid")
	} else if vSize > util.Constants.MaxTxVSize {
		return fmt.Errorf("tx VSize exceeds limit")
	}
	if len(tx.Inputs) > 0 {
		// Not coinbase - verify total outputs < total inputs
		if tx.OutputsValue() >= tx.InputsValue() {
			return fmt.Errorf("tx outputs exceed or match inputs")
		}
	} else {
		// Coinbase - verify outputs exist and total outputs >= BlockReward
		if len(tx.Outputs) != 1 {
			return fmt.Errorf("coinbase must have 1 output")
		} else if tx.OutputsValue() < uint64(util.Constants.BlockReward) {
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
	inv.txs.Store(txId, txRecord{
		tx:    tx,
		vSize: vSize,
	})
	return nil
}
