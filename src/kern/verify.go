package kern

import (
	"fmt"

	"github.com/levilutz/basiccoin/src/util"
)

// Subset of Inv methods needed to verify things.
type InvVerifier interface {
	HasBlock(blockId HashT) bool
	GetBlock(blockId HashT) Block
	GetBlockHeight(blockId HashT) uint64
	HasMerkle(nodeId HashT) bool
	GetMerkleTxIds(root HashT) []HashT
	GetMerkleTxs(root HashT) []Tx
	GetMerkleVSize(merkleId HashT) uint64
	HasTxOut(txId HashT, ind uint64) bool
	GetTxOut(txId HashT, ind uint64) TxOut
	HasEntity(entityId HashT) bool
	GetEntityVSize(entityId HashT) uint64
}

// Encapsualte verification of various entities by the same rules and inv.
type Verifier struct {
	params Params
	inv    InvVerifier
}

// Create a new Verifier.
func NewVerifier(params Params, inv InvVerifier) *Verifier {
	return &Verifier{
		params: params,
		inv:    inv,
	}
}

// Verify a tx.
func (v Verifier) VerifyTx(tx Tx) error {
	if err := v.VerifyTxIsolated(tx); err != nil {
		return err
	}

	// Verify each input's public key and value match origin
	for _, txi := range tx.Inputs {
		if !v.inv.HasTxOut(txi.OriginTxId, txi.OriginTxOutInd) {
			return fmt.Errorf(
				"failed to find utxo %s[%d]",
				txi.OriginTxId,
				txi.OriginTxOutInd,
			)
		}
		origin := v.inv.GetTxOut(txi.OriginTxId, txi.OriginTxOutInd)
		if !DHashBytes(txi.PublicKey).Eq(origin.PublicKeyHash) {
			return fmt.Errorf("given public key does not match claimed utxo")
		}
		if txi.Value != origin.Value {
			return fmt.Errorf("given value does not match claimed utxo")
		}
	}
	return nil
}

// Verify a merkle node.
func (v Verifier) VerifyMerkle(node MerkleNode) error {
	// Verify left child exists and get its size
	totalSize := uint64(0)
	if v.inv.HasEntity(node.LChild) {
		totalSize += v.inv.GetEntityVSize(node.LChild)
	} else {
		return fmt.Errorf("failed to find LChild: %s", node.LChild)
	}

	// Only verify right child if it's different from left child
	if node.RChild != node.LChild {
		// Verify right child exists and get its size
		if v.inv.HasEntity(node.RChild) {
			totalSize += v.inv.GetEntityVSize(node.RChild)
		} else {
			return fmt.Errorf("failed to find RChild: %s", node.RChild)
		}

		// Check no overlap between tx sets of left and right children
		lTxs := util.NewSetFromList(v.inv.GetMerkleTxIds(node.LChild))
		rTxs := util.NewSetFromList(v.inv.GetMerkleTxIds(node.RChild))
		if lTxs.HasIntersection(rTxs) {
			return fmt.Errorf("merkle children are different but share txs")
		}
	}

	// Verify this merkle node doesn't create a tree over size limits
	if totalSize > v.params.MaxBlockVSize {
		return fmt.Errorf("merkle cannot be created - would exceed max block vSize")
	}
	return nil
}

// Verify a block.
func (v Verifier) VerifyBlock(b Block) error {
	if err := v.VerifyBlockIsolated(b); err != nil {
		return err
	}

	// Verify merkle root exists
	if !v.inv.HasMerkle(b.MerkleRoot) {
		return fmt.Errorf("failed to find new block merkle root")
	}

	// Verify parent block exists
	if !v.inv.HasBlock(b.PrevBlockId) {
		return fmt.Errorf("failed to find new block parent id")
	}

	// Verify block has txs (but not too many) and retrieve them
	txs := v.inv.GetMerkleTxs(b.MerkleRoot)
	if len(txs) == 0 {
		return fmt.Errorf("new block has no txs")
	} else if len(txs) > int(BlockMaxTxs(v.params)) {
		return fmt.Errorf("new block has too many txs")
	}

	// Verify coinbase MinBlock is this block's height
	if txs[0].MinBlock != v.inv.GetBlockHeight(b.PrevBlockId)+1 {
		return fmt.Errorf("coinbase MinBlock does not equal height")
	}

	totalInputs := uint64(v.params.BlockReward)
	totalOutputs := uint64(0)
	for i, tx := range txs {
		// Verify first block tx is coinbase
		if i == 0 && !tx.IsCoinbase {
			return fmt.Errorf("first block tx must be coinbase")
		}

		// Verify no other txs are coinbase
		if i != 0 && tx.IsCoinbase {
			return fmt.Errorf("only first block tx may be coinbase")
		}
		totalInputs += tx.InputsValue()
		totalOutputs += tx.OutputsValue()
	}

	// Verify block's total inputs and outputs match
	if totalInputs != totalOutputs {
		return fmt.Errorf("block total inputs and outputs do not match")
	}

	// Verify block total vSize within limits
	if v.inv.GetMerkleVSize(b.MerkleRoot) > v.params.MaxBlockVSize {
		return fmt.Errorf("block exceeds max vSize")
	}
	return nil
}

// Verify what we can about this transaction in isolation.
func (v Verifier) VerifyTxIsolated(tx Tx) error {
	// Verify input signatures match outputs
	preSigHash := TxHashPreSig(tx.MinBlock, tx.Outputs)
	for _, txi := range tx.Inputs {
		valid, err := EcdsaVerify(txi.PublicKey, preSigHash, txi.Signature)
		if err != nil || !valid {
			return fmt.Errorf("tx signature invalid")
		}
	}

	// Verify within vSize limit
	if tx.VSize() > v.params.MaxTxVSize {
		return fmt.Errorf("tx vSize exceeds limit")
	}

	if tx.IsCoinbase {
		// Verify coinbase has no inputs
		if len(tx.Inputs) > 0 {
			return fmt.Errorf("coinbase cannot have inputs")
		}

		// Verify coinbase has only 1 output
		if len(tx.Outputs) != 1 {
			return fmt.Errorf("coinbase must have 1 output")
		}

		// Verify coinbase has at least the minimum block reward
		if tx.OutputsValue() < v.params.BlockReward {
			return fmt.Errorf("coinbase has insufficient block reward")
		}

	} else {
		// Verify non-coinbase has surplus
		if !tx.HasSurplus() {
			return fmt.Errorf("tx outputs exceed or match inputs")
		}
	}
	return nil
}

// Verify what we can about this block in isolation.
func (v Verifier) VerifyBlockIsolated(b Block) error {
	// Verify block hash beats claimed difficulty
	if !b.Hash().Lt(b.Difficulty) {
		return fmt.Errorf("block fails to beat claimed target difficulty")
	}
	return nil
}
