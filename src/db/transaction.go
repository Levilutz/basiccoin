package db

type TxIn struct {
	OriginTxId     HashT
	OriginTxOutInd uint32
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
	Value         uint32
	PublicKeyHash HashT
}

func (txo TxOut) Hash() HashT {
	return DHashItems(txo.Value, txo.PublicKeyHash)
}

type Tx struct {
	MinBlock uint32
	Inputs   []TxIn
	Outputs  []TxOut
}

func (tx Tx) Hash() HashT {
	return DHashItems(
		tx.MinBlock, DHashList(tx.Inputs), DHashList(tx.Outputs),
	)
}

func TxHashPreSig(minBlock uint32, outputs []TxOut) HashT {
	return DHashItems(minBlock, DHashList(outputs))
}
