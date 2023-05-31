package db

type TxIn struct {
	OriginTxId     HashT
	OriginTxOutInd uint64
	PublicKey      []byte
	Signature      []byte
	Value          uint64
}

func (txi TxIn) Hash() HashT {
	return DHashItems(
		txi.OriginTxId[:],
		txi.OriginTxOutInd,
		txi.PublicKey,
		txi.Signature,
		txi.Value,
	)
}

func (txi TxIn) VSize() uint64 {
	// 32 from OriginTxId, 8 from OriginTxOutInd, 8 from Value
	return uint64(32 + 8 + 8 + len(txi.PublicKey) + len(txi.Signature))
}

func TxInPackHasher(txins []TxIn) []Hasher {
	out := make([]Hasher, len(txins))
	for i := 0; i < len(txins); i++ {
		out[i] = txins[i]
	}
	return out
}

type TxOut struct {
	Value         uint64
	PublicKeyHash HashT
}

func (txo TxOut) Hash() HashT {
	return DHashItems(txo.Value, txo.PublicKeyHash)
}

func (txo TxOut) VSize() uint64 {
	// 8 from Value, 32 from PublicKeyHash
	return uint64(8 + 32)
}

type Tx struct {
	MinBlock uint64
	Inputs   []TxIn
	Outputs  []TxOut
}

func (tx Tx) Hash() HashT {
	return DHashItems(
		tx.MinBlock, DHashList(tx.Inputs), DHashList(tx.Outputs),
	)
}

func (tx Tx) InputsValue() uint64 {
	total := uint64(0)
	for _, txi := range tx.Inputs {
		total += uint64(txi.Value)
	}
	return total
}

func (tx Tx) OutputsValue() uint64 {
	total := uint64(0)
	for _, txo := range tx.Outputs {
		total += uint64(txo.Value)
	}
	return total
}

func (tx Tx) VSize() uint64 {
	// 8 from MinBlock, 32 each from top-level hash of Inputs and Outputs
	vSize := uint64(8 + 32 + 32)
	for _, txi := range tx.Inputs {
		vSize += txi.VSize()
	}
	for _, txo := range tx.Outputs {
		vSize += txo.VSize()
	}
	return vSize
}

func (tx Tx) Rate() float64 {
	return float64(tx.InputsValue()-tx.OutputsValue()) / float64(tx.VSize())
}

func (tx Tx) SignaturesValid() bool {
	preSigHash := TxHashPreSig(tx.MinBlock, tx.Outputs)
	for _, txi := range tx.Inputs {
		valid, err := EcdsaVerify(txi.PublicKey, preSigHash, txi.Signature)
		if err != nil || !valid {
			return false
		}
	}
	return true
}

func TxHashPreSig(minBlock uint64, outputs []TxOut) HashT {
	return DHashItems(minBlock, DHashList(outputs))
}

func MinNonCoinbaseVSize() uint64 {
	return Tx{
		MinBlock: 0,
		Inputs: []TxIn{
			{
				OriginTxId:     HashTZero,
				OriginTxOutInd: 0,
				PublicKey:      []byte{},
				Signature:      []byte{},
				Value:          0,
			},
		},
		Outputs: make([]TxOut, 0),
	}.VSize()
}

func CoinbaseVSize() uint64 {
	return Tx{
		MinBlock: 0,
		Inputs:   make([]TxIn, 0),
		Outputs: []TxOut{
			{
				Value:         0,
				PublicKeyHash: HashTZero,
			},
		},
	}.VSize()
}
