package db

type TxIn struct {
	OriginTxId     HashT
	OriginTxOutInd int
	PublicKey      []byte
	Signature      []byte
}

func (txi TxIn) Hash() HashT {
	return HashGenericItems(
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
	return HashGenericItems(txo.Value, txo.PublicKeyHash)
}

type Tx struct {
	MinBlock int
	Inputs   []TxIn
	Outputs  []TxOut
}

func (tx Tx) Hash() HashT {
	return HashGenericItems(
		tx.MinBlock, NewDHashList(tx.Inputs), NewDHashList(tx.Outputs),
	)
}

func HashPreSig(minBlock int, outputs []TxOut) HashT {
	return HashGenericItems(minBlock, NewDHashList(outputs))
}

type Block struct {
	PrevBlockId HashT
	MerkleRoot  HashT
	Difficulty  HashT
	Nonce       HashT
}

func (b Block) Hash() HashT {
	return HashGenericItems(b.PrevBlockId, b.MerkleRoot, b.Difficulty, b.Nonce)
}
