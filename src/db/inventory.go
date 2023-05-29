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
	AncestorDepth(blockId HashT, ancestorId HashT) (uint32, error)
	AnyBlockIdsKnown(blockIds []HashT) (HashT, bool)
	GetBlockHeight(blockId HashT) (uint32, error)
	GetBlockHeritage(blockId HashT, maxLen int) ([]HashT, error)
	GetBlockParentId(blockId HashT) (HashT, error)
	GetTxInOrigin(txi TxIn) (TxOut, error)
	GetTxSurplus(tx Tx) (uint64, error)
	LoadBlock(blockId HashT) (Block, bool)
	LoadMerkle(nodeId HashT) (MerkleNode, bool)
	LoadMerkleTxs(root HashT) ([]Tx, error)
	LoadTx(txId HashT) (Tx, bool)
	LoadTxOrMerkle(id HashT) (*Tx, *MerkleNode)
	VerifyEntityExists(id HashT) error
}

// Write-once read-many maps.
// Only one thread should be making writes at a time, but many can be reading.
type Inv struct {
	blocks  *util.SyncMap[HashT, Block]
	merkles *util.SyncMap[HashT, MerkleNode]
	txs     *util.SyncMap[HashT, Tx]
	blockHs *util.SyncMap[HashT, uint32]
}

func NewInv() *Inv {
	inv := &Inv{
		blocks:  util.NewSyncMap[HashT, Block](),
		merkles: util.NewSyncMap[HashT, MerkleNode](),
		txs:     util.NewSyncMap[HashT, Tx](),
		blockHs: util.NewSyncMap[HashT, uint32](),
	}
	inv.blockHs.Store(HashTZero, 0)
	inv.blocks.Store(HashTZero, Block{})
	return inv
}

// Store a new block, ensures merkle root known and difficulty beat.
func (inv *Inv) StoreBlock(b Block) error {
	blockId := b.Hash()
	if _, ok := inv.LoadBlock(blockId); ok {
		return ErrEntityKnown
	} else if !BelowTarget(blockId, b.Difficulty) {
		return fmt.Errorf("block failed to beat target difficulty")
	}
	txs, err := inv.LoadMerkleTxs(b.MerkleRoot)
	if err != nil {
		return err
	} else if len(txs) == 0 {
		return fmt.Errorf("block has no txs")
	} else if len(txs) > int(util.Constants.MaxBlockTxs) {
		return fmt.Errorf("block has too many txs")
	} else if len(txs[0].Inputs) != 0 {
		return fmt.Errorf("block missing coinbase tx")
	}
	totalInputs := uint64(util.Constants.BlockReward)
	totalOutputs := uint64(0)
	totalVSize := uint64(0)
	for i, tx := range txs {
		if i == 0 {
			totalOutputs += tx.TotalOutputs()
		} else {
			surplus, err := inv.GetTxSurplus(tx)
			if err != nil {
				return err
			}
			totalInputs += surplus
		}
		totalVSize += tx.VSize()
	}
	if totalInputs != totalOutputs {
		return fmt.Errorf("total inputs and outputs do not match")
	} else if totalVSize > util.Constants.MaxBlockVSize {
		return fmt.Errorf("block exceeds max vSize")
	}
	parentHeight, err := inv.GetBlockHeight(b.PrevBlockId)
	if err != nil {
		return err
	}
	inv.blockHs.Store(blockId, parentHeight+1)
	inv.blocks.Store(blockId, b)
	return nil
}

// Load a block, return the block and whether it exists.
func (inv *Inv) LoadBlock(blockId HashT) (Block, bool) {
	return inv.blocks.Load(blockId)
}

// Store a new merkle, ensures children known.
func (inv *Inv) StoreMerkle(merkle MerkleNode) error {
	nodeId := merkle.Hash()
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
func (inv *Inv) StoreTx(tx Tx) error {
	txId := tx.Hash()
	if _, ok := inv.txs.Load(txId); ok {
		return ErrEntityKnown
	}
	if !tx.SignaturesValid() {
		return fmt.Errorf("tx signatures invalid")
	} else if tx.VSize() > util.Constants.MaxTxVSize {
		return fmt.Errorf("tx VSize exceeds limit")
	}
	// Verify outputs < inputs (or outputs >= blockReward if coinbase)
	if len(tx.Inputs) > 0 {
		if _, err := inv.GetTxSurplus(tx); err != nil {
			return err
		}
	} else {
		if tx.TotalOutputs() < uint64(util.Constants.BlockReward) {
			return fmt.Errorf("coinbase has insufficient block reward")
		}
	}
	// Verify given public key matches hash on claimed utxo
	for _, txi := range tx.Inputs {
		origin, err := inv.GetTxInOrigin(txi)
		if err != nil {
			return err
		}
		if DHash(txi.PublicKey) != origin.PublicKeyHash {
			return fmt.Errorf("given public key does not match claimed utxo")
		}
	}
	inv.txs.Store(txId, tx)
	return nil
}

// Load a tx, return the tx and whether it exists.
func (inv *Inv) LoadTx(txId HashT) (Tx, bool) {
	return inv.txs.Load(txId)
}

func (inv *Inv) GetTxInOrigin(txi TxIn) (TxOut, error) {
	originTx, ok := inv.LoadTx(txi.OriginTxId)
	if !ok || txi.OriginTxOutInd >= uint32(len(originTx.Outputs)) {
		return TxOut{}, ErrEntityUnknown
	}
	return originTx.Outputs[txi.OriginTxOutInd], nil
}

func (inv *Inv) GetTxSurplus(tx Tx) (uint64, error) {
	// Sum up total tx inputs
	totalInputs := uint64(0)
	for _, txi := range tx.Inputs {
		origin, err := inv.GetTxInOrigin(txi)
		if err != nil {
			return 0, err
		}
		totalInputs += uint64(origin.Value)
	}
	totalOutputs := tx.TotalOutputs()
	// Verify and return
	if totalOutputs >= totalInputs {
		return 0, fmt.Errorf("tx outputs exceed or match inputs")
	}
	return totalInputs - totalOutputs, nil
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
	if !ok || blockId == HashTZero {
		return HashT{}, ErrEntityUnknown
	}
	return block.PrevBlockId, nil
}

// Get a block's height (0x0 is height 0, origin block is height 1).
func (inv *Inv) GetBlockHeight(blockId HashT) (uint32, error) {
	h, ok := inv.blockHs.Load(blockId)
	if !ok {
		return 0, ErrEntityUnknown
	}
	return h, nil
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

// Load all txs descended from a merkle node.
func (inv *Inv) LoadMerkleTxs(root HashT) ([]Tx, error) {
	outTxs := make([]Tx, 0)
	// Go through each node in tree, categorizing as either tx or merkle
	idQueue := util.NewQueue[HashT]()
	visitedIds := util.NewSet[HashT]() // Prevent cycles
	idQueue.Push(root)
	for i := 0; i < int(util.Constants.MaxTreeSize); i++ {
		// Pop next id, finish if we've cleared queue
		nextId, ok := idQueue.Pop()
		if !ok {
			break
		}

		// Prevent cycles
		if visitedIds.Includes(nextId) {
			return nil, fmt.Errorf("circular visit to id %x", nextId)
		}
		visitedIds.Add(nextId)

		// Load tx or merkle and categorize
		tx, merkle := inv.LoadTxOrMerkle(nextId)
		if tx != nil {
			outTxs = append(outTxs, *tx)
		} else if merkle != nil {
			idQueue.Push(merkle.LChild)
			if merkle.RChild != merkle.LChild {
				idQueue.Push(merkle.RChild)
			}
		} else {
			return nil, ErrEntityUnknown
		}
	}

	// Verify we didn't just hit limit
	_, ok := idQueue.Pop()
	if ok {
		return nil, fmt.Errorf(
			"tree exceeds max size of %d", util.Constants.MaxTreeSize,
		)
	}
	return outTxs, nil
}

// Store full block with any new merkle nodes and txs. Only merkles / txs reachable
// from the block merkleRoot are included, missing merkles and txs cause failure.
func (inv *Inv) StoreFullBlock(
	block Block, merkles []MerkleNode, txs []Tx,
) error {
	blockId := block.Hash()
	// Skip if known
	_, ok := inv.LoadBlock(blockId)
	if ok {
		return ErrEntityKnown
	}

	merkleMap := HasherMap(merkles)
	txMap := HasherMap(txs)

	// What to add at the end
	newMerkles := make([]MerkleNode, 0)
	newTxs := make([]Tx, 0)

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
		if tx, ok := txMap[nextId]; ok {
			// Id exists as tx in args
			newTxs = append(newTxs, tx)
		} else if merkle, ok := merkleMap[nextId]; ok {
			// Id exists as merkle in args
			newMerkles = append(newMerkles, merkle)
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
	for _, tx := range newTxs {
		err := inv.StoreTx(tx)
		if err != nil {
			return err
		}
	}
	for _, merkle := range newMerkles {
		err := inv.StoreMerkle(merkle)
		if err != nil {
			return err
		}
	}
	err := inv.StoreBlock(block)
	if err != nil {
		return err
	}
	return nil
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
