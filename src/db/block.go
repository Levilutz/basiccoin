package db

import (
	"fmt"

	"github.com/levilutz/basiccoin/src/util"
)

type MerkleNode struct {
	LChild HashT
	RChild HashT
}

func (node MerkleNode) Hash() HashT {
	return DHashItems(node.LChild, node.RChild)
}

type Block struct {
	PrevBlockId HashT
	MerkleRoot  HashT
	Difficulty  HashT
	Nonce       uint64
}

func (b Block) Hash() HashT {
	return DHashItems(b.PrevBlockId, b.MerkleRoot, b.Difficulty, b.Nonce)
}

// Verify that the claimed proof of work is valid.
func (b Block) VerifyProofOfWork() error {
	if !BelowTarget(b.Hash(), b.Difficulty) {
		return fmt.Errorf("failed to beat claimed target")
	}
	return nil
}

// The maximum number of txs that could theoretically be in a block, including coinbase.
func BlockMaxTxs() uint64 {
	standardTxSpace := util.Constants.MaxBlockVSize - CoinbaseVSize()
	// +1 to "round up"
	maxStandardTxs := standardTxSpace/MinNonCoinbaseVSize() + 1
	// +1 to re-include coinbase tx
	return maxStandardTxs + 1
}

// The maximum possible size of a block's merkle tree, including tx leafs.
func MerkleTreeMaxSize() uint64 {
	return BlockMaxTxs()*2 - 1
}
