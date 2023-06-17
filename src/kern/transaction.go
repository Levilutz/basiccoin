package kern

// Reference to unspent transaction output.
// This is just a subset of the fields in a TxIn.
type Utxo struct {
	TxId  HashT  `json:"txId"`
	Ind   uint64 `json:"ind"`
	Value uint64 `json:"value"`
}

func UtxoFromInput(txi TxIn) Utxo {
	return Utxo{
		TxId:  txi.OriginTxId,
		Ind:   txi.OriginTxOutInd,
		Value: txi.Value,
	}
}

// A transaction input.
type TxIn struct {
	OriginTxId     HashT  `json:"originTxId"`
	OriginTxOutInd uint64 `json:"originTxOutInd"`
	PublicKey      []byte `json:"publicKey"`
	Signature      []byte `json:"signature"`
	Value          uint64 `json:"value"`
}

func (txi TxIn) Hash() HashT {
	return DHashVarious(
		txi.OriginTxId,
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

// A transaction output.
type TxOut struct {
	Value         uint64 `json:"value"`
	PublicKeyHash HashT  `json:"publicKeyHash"`
}

func (txo TxOut) Hash() HashT {
	return DHashVarious(txo.Value, txo.PublicKeyHash)
}

func (txo TxOut) VSize() uint64 {
	// 8 from Value, 32 from PublicKeyHash
	return uint64(8 + 32)
}

// A transaction.
type Tx struct {
	IsCoinbase bool    `json:"isCoinbase"`
	MinBlock   uint64  `json:"minBlock"`
	Inputs     []TxIn  `json:"inputs"`
	Outputs    []TxOut `json:"outputs"`
}

func (tx Tx) Hash() HashT {
	return DHashVarious(
		tx.IsCoinbase, tx.MinBlock, DHashList(tx.Inputs), DHashList(tx.Outputs),
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
	// Float inputs and outputs separately so we don't uint underflow
	return (float64(tx.InputsValue()) - float64(tx.OutputsValue())) / float64(tx.VSize())
}

func (tx Tx) HasSurplus() bool {
	return tx.InputsValue() > tx.OutputsValue()
}

func (tx Tx) GetConsumedUtxos() []Utxo {
	out := make([]Utxo, len(tx.Inputs))
	for i := range out {
		out[i] = UtxoFromInput(tx.Inputs[i])
	}
	return out
}

func TxHashPreSig(minBlock uint64, outputs []TxOut) HashT {
	return DHashVarious(minBlock, DHashList(outputs))
}

func MinNonCoinbaseVSize() uint64 {
	return Tx{
		MinBlock: 0,
		Inputs: []TxIn{
			{
				OriginTxId:     HashT{},
				OriginTxOutInd: 0,
				PublicKey:      ExamplePubDer(),
				Signature:      []byte{}, // No lower bound on signature length
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
				PublicKeyHash: HashT{},
			},
		},
	}.VSize()
}
