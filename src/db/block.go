package db

import "fmt"

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
	Nonce       uint32
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
