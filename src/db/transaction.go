package db

import "fmt"

type TxIn struct {
	OriginTxId     HashT
	OriginTxOutInd int
	PublicKey      []byte
	Signature      []byte
}

func (txi TxIn) Hash() HashT {
	return DHashItems(
		txi.OriginTxId[:],
		txi.OriginTxOutInd,
		txi.PublicKey,
		txi.Signature,
	)
}

func TxInPackHasher(txins []TxIn) []Hasher {
	out := make([]Hasher, len(txins))
	for i := 0; i < len(txins); i++ {
		out[i] = txins[i]
	}
	return out
}

type TxOut struct {
	Value         int
	PublicKeyHash HashT
}

func (txo TxOut) Hash() HashT {
	return DHashItems(txo.Value, txo.PublicKeyHash)
}

type Tx struct {
	MinBlock int
	Inputs   []TxIn
	Outputs  []TxOut
}

func (tx Tx) Hash() HashT {
	return DHashItems(
		tx.MinBlock, DHashList(tx.Inputs), DHashList(tx.Outputs),
	)
}

func TxHashPreSig(minBlock int, outputs []TxOut) HashT {
	return DHashItems(minBlock, DHashList(outputs))
}

type BlockHeader struct {
	PrevBlockId HashT
	MerkleRoot  HashT
	Difficulty  HashT
	Nonce       HashT
}

func (bh BlockHeader) Hash() HashT {
	return DHashItems(bh.PrevBlockId, bh.MerkleRoot, bh.Difficulty, bh.Nonce)
}

func (bh BlockHeader) Verify() error {
	// Verify hash matches claimed target difficulty
	blockHash := bh.Hash()
	if !BelowTarget(blockHash, bh.Difficulty) {
		return fmt.Errorf("block does not beat claimed difficulty")
	}

	return nil
}

func (bh BlockHeader) VerifyMerkle(txIds []HashT) error {
	// Verify merkle root matches
	merkleRoot := DHashHashes(txIds)
	if merkleRoot != bh.MerkleRoot {
		return fmt.Errorf("invalid claimed merkle root: %s", bh.MerkleRoot)
	}

	return nil
}
