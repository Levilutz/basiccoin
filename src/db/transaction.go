package db

import (
	"encoding/json"
	"fmt"
)

// Reference to unspent transaction output.
// This is just a subset of the fields in a TxIn.
type Utxo struct {
	TxId  HashT
	Ind   uint64
	Value uint64
}

type UtxoJSON struct {
	TxId  string `json:"txId"`
	Ind   uint64 `json:"ind"`
	Value uint64 `json:"value"`
}

func (utxo Utxo) MarshalJSON() ([]byte, error) {
	return json.Marshal(UtxoJSON{
		TxId:  fmt.Sprintf("%x", utxo.TxId),
		Ind:   utxo.Ind,
		Value: utxo.Value,
	})
}

func (utxo *Utxo) UnmarshalJSON(data []byte) error {
	v := UtxoJSON{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	txId, err := StringToHash(v.TxId)
	if err != nil {
		return err
	}
	utxo.TxId = txId
	utxo.Ind = v.Ind
	utxo.Value = v.Value
	return nil
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
	OriginTxId     HashT
	OriginTxOutInd uint64
	PublicKey      []byte
	Signature      []byte
	Value          uint64
}

type TxInJSON struct {
	OriginTxId     string `json:"originTxId"`
	OriginTxOutInd uint64 `json:"originTxOutInd"`
	PublicKey      []byte `json:"publicKey"`
	Signature      []byte `json:"signature"`
	Value          uint64 `json:"value"`
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

func (txi TxIn) MarshalJSON() ([]byte, error) {
	return json.Marshal(TxInJSON{
		OriginTxId:     fmt.Sprintf("%x", txi.OriginTxId),
		OriginTxOutInd: txi.OriginTxOutInd,
		PublicKey:      txi.PublicKey,
		Signature:      txi.Signature,
		Value:          txi.Value,
	})
}

func (txi *TxIn) UnmarshalJSON(data []byte) error {
	v := TxInJSON{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	origin, err := StringToHash(v.OriginTxId)
	if err != nil {
		return err
	}
	txi.OriginTxId = origin
	txi.OriginTxOutInd = v.OriginTxOutInd
	txi.PublicKey = v.PublicKey
	txi.Signature = v.Signature
	txi.Value = v.Value
	return nil
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

// A transaction output.
type TxOut struct {
	Value         uint64
	PublicKeyHash HashT
}

type TxOutJSON struct {
	Value         uint64 `json:"value"`
	PublicKeyHash string `json:"publicKeyHash"`
}

func (txo TxOut) Hash() HashT {
	return DHashItems(txo.Value, txo.PublicKeyHash)
}

func (txo TxOut) MarshalJSON() ([]byte, error) {
	return json.Marshal(TxOutJSON{
		Value:         txo.Value,
		PublicKeyHash: fmt.Sprintf("%x", txo.PublicKeyHash),
	})
}

func (txo *TxOut) UnmarshalJSON(data []byte) error {
	v := TxOutJSON{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	pkh, err := StringToHash(v.PublicKeyHash)
	if err != nil {
		return err
	}
	txo.Value = v.Value
	txo.PublicKeyHash = pkh
	return nil
}

func (txo TxOut) VSize() uint64 {
	// 8 from Value, 32 from PublicKeyHash
	return uint64(8 + 32)
}

// A transaction.
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
	// Float inputs and outputs separately so we don't uint underflow
	return (float64(tx.InputsValue()) - float64(tx.OutputsValue())) / float64(tx.VSize())
}

func (tx Tx) HasSurplus() bool {
	return tx.InputsValue() > tx.OutputsValue()
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

func (tx Tx) GetConsumedUtxos() []Utxo {
	out := make([]Utxo, len(tx.Inputs))
	for i := range out {
		out[i] = UtxoFromInput(tx.Inputs[i])
	}
	return out
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
