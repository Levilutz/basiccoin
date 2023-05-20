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

func HashPreSig(minBlock int, outputs []TxOut) HashT {
	return DHashItems(minBlock, DHashList(outputs))
}

type Block struct {
	PrevBlockId HashT
	MerkleRoot  HashT
	Difficulty  HashT
	Nonce       HashT
	TxIds       []HashT
}

func (b Block) Hash() HashT {
	return DHashItems(b.PrevBlockId, b.MerkleRoot, b.Difficulty, b.Nonce)
}

func (b Block) VerifyInternal() error {
	// Verify merkle root matches
	merkleRoot := DHashHashes(b.TxIds)
	if merkleRoot != b.MerkleRoot {
		return fmt.Errorf("invalid claimed merkle root: %s", b.MerkleRoot)
	}

	// Verify hash matches claimed target difficulty
	blockHash := b.Hash()
	if !BelowTarget(blockHash, b.Difficulty) {
		return fmt.Errorf("block does not beat claimed difficulty")
	}

	return nil
}
