package db

import (
	"fmt"

	"github.com/levilutz/basiccoin/src/kern"
	"github.com/levilutz/basiccoin/src/util"
)

// Verify various things about a candidate new block.
// Checks block hash beats claimed target difficulty.
// Checks merkle root exists.
// Checks parent block exists.
// Checks 0 < num txs <= max allowed txs.
// Checks coinbase MinBlock matches height.
// CHecks coinbase has no inputs and 1 sufficient output.
// Checks non-coinbase txs have surplus inputs.
// Checks total of all txs (including coinbase) has 0 surplus.
// Checks block vSize is within limit (covered by verifyMerkle, but just to be safe).
func verifyBlock(inv InvReader, block kern.Block) error {
	if err := block.VerifyIsolated(); err != nil {
		return err
	}
	if !inv.HasMerkle(block.MerkleRoot) {
		return fmt.Errorf("failed to find new block merkle root")
	} else if !inv.HasBlock(block.PrevBlockId) {
		return fmt.Errorf("failed to find new block parent id")
	}
	parentHeight := inv.GetBlockHeight(block.PrevBlockId)
	txs := inv.GetMerkleTxs(block.MerkleRoot)
	if len(txs) == 0 {
		return fmt.Errorf("new block has no txs")
	} else if len(txs) > int(kern.BlockMaxTxs()) {
		return fmt.Errorf("new block has too many txs")
	} else if txs[0].MinBlock != parentHeight+1 {
		return fmt.Errorf("coinbase MinBlock does not equal height")
	}
	totalInputs := uint64(util.Constants.BlockReward)
	totalOutputs := uint64(0)
	for i, tx := range txs {
		if i == 0 && !tx.IsCoinbase {
			return fmt.Errorf("first block tx must be coinbase")
		} else if i != 0 && tx.IsCoinbase {
			return fmt.Errorf("only first block tx may be coinbase")
		}
		totalInputs += tx.InputsValue()
		totalOutputs += tx.OutputsValue()
	}
	if totalInputs != totalOutputs {
		return fmt.Errorf("block total inputs and outputs do not match")
	} else if inv.GetMerkleVSize(block.MerkleRoot) > util.Constants.MaxBlockVSize {
		return fmt.Errorf("block exceeds max vSize")
	}
	return nil
}

// Verify various things about a candidate new merkle node.
// Checks left child exists.
// Checks right child exists, if different from left child.
// Checks new total vSize is within limit.
// Checks no overlap between tx sets of children.
func verifyMerkle(inv InvReader, merkle kern.MerkleNode) error {
	// Get left child size
	totalSize := uint64(0)
	if inv.HasEntity(merkle.LChild) {
		totalSize += inv.GetEntityVSize(merkle.LChild)
	} else {
		return fmt.Errorf("failed to find LChild: %s", merkle.LChild)
	}
	// Get right child size (if appropriate)
	if merkle.RChild != merkle.LChild {
		if inv.HasEntity(merkle.RChild) {
			totalSize += inv.GetEntityVSize(merkle.RChild)
		} else {
			return fmt.Errorf("failed to find RChild: %s", merkle.RChild)
		}
		// Check no overlap between tx sets of children
		lTxs := util.NewSetFromList(inv.GetMerkleTxIds(merkle.LChild))
		rTxs := util.NewSetFromList(inv.GetMerkleTxIds(merkle.RChild))
		if lTxs.HasIntersection(rTxs) {
			return fmt.Errorf("merkle children are different but share txs")
		}
	}
	if totalSize > util.Constants.MaxBlockVSize {
		return fmt.Errorf("merkle cannot be created - would exceed max block vSize")
	}
	return nil
}

// Verify various things about a candidate new tx.
// Checks tx signatures valid over outputs.
// Checks tx vSize within limit.
// Checks tx outputs have surplus if tx does not seem to be coinbase.
// Checks one sufficient output if tx seems to be coinbase.
// Checks claimed utxos exist for each tx input.
// Checks tx public key matches hash on claimed utxo for each tx input.
// Checks tx input value matches utxo value for each tx input.
func verifyTx(inv InvReader, tx kern.Tx) error {
	if err := tx.VerifyIsolated(); err != nil {
		return err
	}
	// Verify given public key and value match those on origin utxo
	for _, txi := range tx.Inputs {
		if !inv.HasTxOut(txi.OriginTxId, txi.OriginTxOutInd) {
			return fmt.Errorf(
				"failed to find utxo %s[%d]",
				txi.OriginTxId,
				txi.OriginTxOutInd,
			)
		}
		origin := inv.GetTxOut(txi.OriginTxId, txi.OriginTxOutInd)
		if !kern.DHashBytes(txi.PublicKey).Eq(origin.PublicKeyHash) {
			return fmt.Errorf("given public key does not match claimed utxo")
		}
		if txi.Value != origin.Value {
			return fmt.Errorf("given value does not match claimed utxo")
		}
	}
	return nil
}
