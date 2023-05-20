package db

type TxIn struct {
	OriginTxId     HashT
	OriginTxOutInd int
	PublicKey      []byte
	Signature      []byte
}

func (txi TxIn) Hash() HashT {
	originTxIdHash := NewDHash(txi.OriginTxId[:])
	originTxOutIndHash := NewDHashInt(txi.OriginTxOutInd)
	publicKeyHash := NewDHash(txi.PublicKey)
	signatureHash := NewDHash(txi.Signature)
	return NewDHash(
		originTxIdHash[:],
		originTxOutIndHash[:],
		publicKeyHash[:],
		signatureHash[:],
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
	valueHash := NewDHashInt(txo.Value)
	return NewDHash(valueHash[:], txo.PublicKeyHash[:])
}

type Tx struct {
	MinBlock int
	Inputs   []TxIn
	Outputs  []TxOut
}

func (tx Tx) Hash() HashT {
	minBlockHash := NewDHashInt(tx.MinBlock)
	inputsHash := NewDHashList(tx.Inputs)
	outputsHash := NewDHashList(tx.Outputs)
	return NewDHash(minBlockHash[:], inputsHash[:], outputsHash[:])
}

func HashPreSig(minBlock int, outputs []TxOut) HashT {
	minBlockHash := NewDHashInt(minBlock)
	outputsHash := NewDHashList(outputs)
	return NewDHash(minBlockHash[:], outputsHash[:])
}
