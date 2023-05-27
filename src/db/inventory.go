package db

import (
	"fmt"

	"github.com/levilutz/basiccoin/src/util"
)

type InvReader interface {
	LoadBlock(blockId HashT) (Block, bool)
	LoadMerkle(nodeId HashT) (MerkleNode, bool)
	LoadTx(txId HashT) (Tx, bool)
	LoadTxOrMerkle(id HashT) (*Tx, *MerkleNode)
	LoadFullBlock(blockId HashT) (Block, map[HashT]MerkleNode, map[HashT]Tx, error)
	VerifyEntity(id HashT) error
}

// Write-once maps
type Inv struct {
	blocks  util.SyncMap[HashT, Block]
	merkles util.SyncMap[HashT, MerkleNode]
	txs     util.SyncMap[HashT, Tx]
}

// Store a new block, ensures merkle root known.
func (inv *Inv) StoreBlock(blockId HashT, b Block) error {
	if _, ok := inv.blocks.Load(blockId); ok {
		return fmt.Errorf("block known")
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
		return fmt.Errorf("merkle known")
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
		return fmt.Errorf("tx known")
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

// TODO: Check StoreXXXs for error here and in LoadFullBlock
// Store full block with any new merkle nodes and txs. Only merkles / txs reachable
// from the block merkleRoot are included, missing merkles and txs cause failure.
func (inv *Inv) StoreFullBlock(
	blockId HashT, block Block, merkles map[HashT]MerkleNode, txs map[HashT]Tx,
) error {
	// Skip if known
	_, ok := inv.LoadBlock(blockId)
	if ok {
		return fmt.Errorf("block known")
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
			return fmt.Errorf("unrecognized entity not given %x", nextId)
		}
	}

	// Add from the sets created earlier
	for txId, tx := range newTxs {
		inv.StoreTx(txId, tx)
	}
	for merkleId, merkle := range newMerkles {
		inv.StoreMerkle(merkleId, merkle)
	}
	inv.StoreBlock(blockId, block)
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
		return Block{}, nil, nil, fmt.Errorf("unknown block: %x", blockId)
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
			return Block{}, nil, nil, fmt.Errorf("unrecognized entity %x", nextId)
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
		return fmt.Errorf("unrecognized entity %x", id)
	}
	return nil
}
