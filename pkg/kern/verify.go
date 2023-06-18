package kern

import (
	"fmt"
	"sort"
	"time"

	"github.com/levilutz/basiccoin/pkg/set"
)

// Subset of Inv methods needed to verify things.
type InvVerifier interface {
	HasBlock(blockId HashT) bool
	GetBlock(blockId HashT) Block
	GetBlockHeight(blockId HashT) uint64
	GetBlockAncestors(blockId HashT, maxLen int) []HashT
	GetBlockSpecificAncestor(blockId HashT, depth int) HashT
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
		if !v.inv.HasTxOut(txi.Utxo.TxId, txi.Utxo.Ind) {
			return fmt.Errorf("failed to find utxo %s[%d]", txi.Utxo.TxId, txi.Utxo.Ind)
		}
		origin := v.inv.GetTxOut(txi.Utxo.TxId, txi.Utxo.Ind)
		if !DHashBytes(txi.PublicKey).Eq(origin.PublicKeyHash) {
			return fmt.Errorf("given public key does not match claimed utxo")
		}
		if txi.Utxo.Value != origin.Value {
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
		lTxs := set.NewSetFromList(v.inv.GetMerkleTxIds(node.LChild))
		rTxs := set.NewSetFromList(v.inv.GetMerkleTxIds(node.RChild))
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

	newBlockHeight := v.inv.GetBlockHeight(b.PrevBlockId) + 1

	// Verify coinbase MinBlock is this block's height
	if txs[0].MinBlock != newBlockHeight {
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

	// Verify block mined time is above median of previous 5 (if not the first)
	if !b.PrevBlockId.EqZero() {
		// Get last 11 blocks, dropping the zero block if appropriate
		ancestorIds := []HashT{b.PrevBlockId}
		ancestorIds = append(ancestorIds, v.inv.GetBlockAncestors(b.PrevBlockId, 4)...)
		if ancestorIds[len(ancestorIds)-1].EqZero() {
			ancestorIds = ancestorIds[:len(ancestorIds)-1]
		}

		// Get median time
		times := make([]uint64, len(ancestorIds))
		for i, blockId := range ancestorIds {
			times[i] = v.inv.GetBlock(blockId).MinedTime
		}
		sort.Slice(times, func(i, j int) bool { return times[i] < times[j] })
		var median uint64
		if len(times)%2 == 0 {
			median = (times[len(times)/2-1] + times[len(times)/2]) / 2
		} else {
			median = times[(len(times)-1)/2]
		}

		// Verify
		if b.MinedTime <= median {
			fmt.Println(uint64(time.Now().Unix()), times, b.MinedTime)
			return fmt.Errorf("block mined time not above median of previous 5")
		}
	}

	// If last block in period, verify time is ahead of first block in period
	// This prevents panics in ExpectedTargetAdjustment
	if newBlockHeight+1%v.params.DifficultyPeriod == 0 {
		var firstBlockId HashT
		if newBlockHeight+1 == v.params.DifficultyPeriod {
			if !v.inv.GetBlockSpecificAncestor(b.PrevBlockId, int(v.params.DifficultyPeriod-2)).EqZero() {
				panic("ancestor state unexpected - should be unreachable")
			}
			firstBlockId = v.inv.GetBlockSpecificAncestor(b.PrevBlockId, int(v.params.DifficultyPeriod-3))
		} else {
			firstBlockId = v.inv.GetBlockSpecificAncestor(b.PrevBlockId, int(v.params.DifficultyPeriod-2))
		}
		firstBlockTime := v.inv.GetBlock(firstBlockId).MinedTime
		if b.MinedTime <= firstBlockTime {
			return fmt.Errorf("block is last in period, yet was mined before first in period")
		}
	}

	// Verify block mined time less than an hour in the future
	if b.MinedTime > uint64(time.Now().Unix())+3600 {
		return fmt.Errorf("block mined time more than one hour in the future")
	}

	// Verify block target adjustment correct
	if !b.PrevBlockId.EqZero() {
		prevTarget := v.inv.GetBlock(b.PrevBlockId).Target
		if newBlockHeight%v.params.DifficultyPeriod == 0 {
			// Verify new target isn't too hard compared to the last
			if b.Target.Lt(prevTarget.MinNextTarget()) {
				return fmt.Errorf("block target reduced more than 4x")
			}

			// Verify new target isn't too easy compared to the last
			if prevTarget.MaxNextTarget(v.params).Lt(b.Target) {
				return fmt.Errorf("block target increased more than 4x")
			}

		} else {
			// Verify target unchanged
			if !b.Target.Eq(prevTarget) {
				return fmt.Errorf("block altering target from parent out of period")
			}
		}
	} else {
		// This is first block - verify target correct
		if !b.Target.Eq(v.params.OriginalTarget) {
			return fmt.Errorf("first block does not have required target")
		}
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
	// Verify block hash beats claimed target
	if !b.Hash().Lt(b.Target) {
		return fmt.Errorf("block fails to beat claimed target target")
	}
	return nil
}
