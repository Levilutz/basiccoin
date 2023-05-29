package db

import (
	"errors"
	"fmt"

	"github.com/levilutz/basiccoin/src/util"
)

var ErrEntityKnown = errors.New("entity known")
var ErrEntityUnknown = errors.New("entity unknown")

type InvReader interface {
	LoadBlock(blockId HashT) (Block, bool)
	LoadMerkle(nodeId HashT) (MerkleNode, bool)
	LoadTx(txId HashT) (Tx, bool)
	LoadTxOrMerkle(id HashT) (*Tx, *MerkleNode)
	GetBlockParentId(blockId HashT) (HashT, error)
	GetBlockHeritage(blockId HashT, maxLen int) ([]HashT, error)
	AnyBlockIdsKnown(blockIds []HashT) (HashT, bool)
	LoadFullBlock(blockId HashT) (Block, map[HashT]MerkleNode, map[HashT]Tx, error)
	VerifyEntityExists(id HashT) error
	AncestorDepth(blockId, ancestorId HashT) (uint32, error)
}

// Write-once read-many maps.
// Only one thread should be making writes at a time, but many can be reading.
type Inv struct {
	blocks  *util.SyncMap[HashT, Block]
	merkles *util.SyncMap[HashT, MerkleNode]
	txs     *util.SyncMap[HashT, Tx]
}

func NewInv() *Inv {
	return &Inv{
		blocks:  util.NewSyncMap[HashT, Block](),
		merkles: util.NewSyncMap[HashT, MerkleNode](),
		txs:     util.NewSyncMap[HashT, Tx](),
	}
}

// Store a new block, ensures merkle root known.
func (inv *Inv) StoreBlock(blockId HashT, b Block) error {
	if _, ok := inv.blocks.Load(blockId); ok {
		return ErrEntityKnown
	} else if err := inv.VerifyEntityExists(b.MerkleRoot); err != nil {
		return err
	}
	inv.blocks.Store(blockId, b)
	return nil
}

// Load a block, return the block and whether it exists.
func (inv *Inv) LoadBlock(blockId HashT) (Block, bool) {
	return inv.blocks.Load(blockId)
}

// Store a new merkle, ensures children known.
func (inv *Inv) StoreMerkle(nodeId HashT, merkle MerkleNode) error {
	if _, ok := inv.merkles.Load(nodeId); ok {
		return ErrEntityKnown
	} else if err := inv.VerifyEntityExists(merkle.LChild); err != nil {
		return err
	} else if err := inv.VerifyEntityExists(merkle.RChild); err != nil {
		return err
	}
	inv.merkles.Store(nodeId, merkle)
	return nil
}

// Load a merkle node, return the merkle and whether it exists.
func (inv *Inv) LoadMerkle(nodeId HashT) (MerkleNode, bool) {
	return inv.merkles.Load(nodeId)
}

// Store a new transaction.
func (inv *Inv) StoreTx(txId HashT, tx Tx) error {
	if _, ok := inv.txs.Load(txId); ok {
		return ErrEntityKnown
	}
	inv.txs.Store(txId, tx)
	return nil
}

// Load a tx, return the tx and whether it exists.
func (inv *Inv) LoadTx(txId HashT) (Tx, bool) {
	return inv.txs.Load(txId)
}

// Load a tx or merkle, return a pointer to whichever exists and nil.
func (inv *Inv) LoadTxOrMerkle(id HashT) (*Tx, *MerkleNode) {
	merkle, ok := inv.LoadMerkle(id)
	if ok {
		return nil, &merkle
	}
	tx, ok := inv.LoadTx(id)
	if ok {
		return &tx, nil
	}
	return nil, nil
}

func (inv *Inv) GetBlockParentId(blockId HashT) (HashT, error) {
	block, ok := inv.LoadBlock(blockId)
	if !ok {
		return HashT{}, ErrEntityUnknown
	}
	return block.PrevBlockId, nil
}

func (inv *Inv) GetBlockHeritage(blockId HashT, maxLen int) ([]HashT, error) {
	out := make([]HashT, 0)
	next := blockId
	var err error
	for i := 0; i < maxLen; i++ {
		next, err = inv.GetBlockParentId(next)
		if err != nil {
			return nil, err
		}
		out = append(out, next)
		if next == HashTZero {
			break
		}
	}
	return out, nil
}

func (inv *Inv) AnyBlockIdsKnown(blockIds []HashT) (HashT, bool) {
	for i := 0; i < len(blockIds); i++ {
		_, ok := inv.LoadBlock(blockIds[i])
		if ok {
			return blockIds[i], true
		}
	}
	return HashTZero, false
}

// Store full block with any new merkle nodes and txs. Only merkles / txs reachable
// from the block merkleRoot are included, missing merkles and txs cause failure.
func (inv *Inv) StoreFullBlock(
	blockId HashT, block Block, merkles map[HashT]MerkleNode, txs map[HashT]Tx,
) error {
	// Skip if known
	_, ok := inv.LoadBlock(blockId)
	if ok {
		return ErrEntityKnown
	}

	// What to add at the end
	newMerkles := make(map[HashT]MerkleNode)
	newTxs := make(map[HashT]Tx)

	// Go through tree from merkle root. If unknown but provided in args, add to inv
	idQueue := util.NewQueue[HashT]()
	visitedIds := util.NewSet[HashT]() // Prevent cycles
	idQueue.Push(block.MerkleRoot)
	for i := 0; i < int(util.Constants.MaxTreeSize); i++ {
		// Pop next id, finish if we've cleared queue
		nextId, ok := idQueue.Pop()
		if !ok {
			break
		}

		// Prevent cycles
		if visitedIds.Includes(nextId) {
			return fmt.Errorf("circular visit to id %x", nextId)
		}
		visitedIds.Add(nextId)

		// If entity known, skip
		txP, merkleP := inv.LoadTxOrMerkle(nextId)
		if txP != nil || merkleP != nil {
			continue
		}

		// Unknown, check if id exists as merkle or tx in args
		if tx, ok := txs[nextId]; ok {
			// Id exists as tx in args
			newTxs[nextId] = tx
		} else if merkle, ok := merkles[nextId]; ok {
			// Id exists as merkle in args
			newMerkles[nextId] = merkle
			idQueue.Push(merkle.LChild)
			if merkle.RChild != merkle.LChild {
				idQueue.Push(merkle.RChild)
			}
		} else {
			// Unrecognized and not provided in args, err
			return ErrEntityUnknown
		}
	}

	// Add from the sets created earlier
	for txId, tx := range newTxs {
		err := inv.StoreTx(txId, tx)
		if err != nil {
			return err
		}
	}
	for merkleId, merkle := range newMerkles {
		err := inv.StoreMerkle(merkleId, merkle)
		if err != nil {
			return err
		}
	}
	err := inv.StoreBlock(blockId, block)
	if err != nil {
		return err
	}
	return nil
}

// Given a block id, load all merkle nodes and transactions from the block.
func (inv *Inv) LoadFullBlock(
	blockId HashT,
) (Block, map[HashT]MerkleNode, map[HashT]Tx, error) {
	outMerkles := make(map[HashT]MerkleNode)
	outTxs := make(map[HashT]Tx)

	// Retrieve block header
	b, ok := inv.LoadBlock(blockId)
	if !ok {
		return Block{}, nil, nil, ErrEntityUnknown
	}

	// Go through each node in tree, categorizing as either tx or merkle
	idQueue := util.NewQueue[HashT]()
	visitedIds := util.NewSet[HashT]() // Prevent cycles
	idQueue.Push(b.MerkleRoot)
	for i := 0; i < int(util.Constants.MaxTreeSize); i++ {
		// Pop next id, finish if we've cleared queue
		nextId, ok := idQueue.Pop()
		if !ok {
			break
		}

		// Prevent cycles
		if visitedIds.Includes(nextId) {
			return Block{}, nil, nil, fmt.Errorf("circular visit to id %x", nextId)
		}
		visitedIds.Add(nextId)

		// Load tx or merkle and categorize
		tx, merkle := inv.LoadTxOrMerkle(nextId)
		if tx != nil {
			outTxs[nextId] = *tx
		} else if merkle != nil {
			outMerkles[nextId] = *merkle
			idQueue.Push(merkle.LChild)
			if merkle.RChild != merkle.LChild {
				idQueue.Push(merkle.RChild)
			}
		} else {
			return Block{}, nil, nil, ErrEntityUnknown
		}
	}

	// Verify we didn't just hit limit
	_, ok = idQueue.Pop()
	if ok {
		return Block{}, nil, nil, fmt.Errorf(
			"tree exceeds max size of %d", util.Constants.MaxTreeSize,
		)
	}

	return b, outMerkles, outTxs, nil
}

// Returns error if entity not known, or nil if known.
func (inv *Inv) VerifyEntityExists(id HashT) error {
	txP, merkleP := inv.LoadTxOrMerkle(id)
	if txP == nil && merkleP == nil {
		return ErrEntityUnknown
	}
	return nil
}

// Returns how many blocks deep the ancestor is.
func (inv *Inv) AncestorDepth(blockId, ancestorId HashT) (uint32, error) {
	var err error
	depth := uint32(0)
	for blockId != ancestorId && blockId != HashTZero {
		blockId, err = inv.GetBlockParentId(blockId)
		if err != nil {
			return 0, err
		}
	}
	if blockId != ancestorId {
		return 0, fmt.Errorf("block does not trace to ancestor")
	}
	return depth, nil
}
