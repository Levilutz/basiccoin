package db_test

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/json"
	"testing"

	"github.com/levilutz/basiccoin/src/db"
	. "github.com/levilutz/basiccoin/src/db"
	"github.com/levilutz/basiccoin/src/util"
)

// Test that a transaction can be hashed in full.
func TestTransactionHash(t *testing.T) {
	// Generate signatures
	generateWithPublic := func() (*ecdsa.PrivateKey, []byte) {
		priv, err := NewEcdsa()
		util.AssertNoErr(t, err)
		pubDer, err := MarshalEcdsaPublic(priv)
		util.AssertNoErr(t, err)
		return priv, pubDer
	}
	inKey1Priv, inKey1PubDer := generateWithPublic()
	inKey2Priv, inKey2PubDer := generateWithPublic()
	_, outKey1PubDer := generateWithPublic()
	_, outKey2PubDer := generateWithPublic()
	inKey1PrivDer, err := MarshalEcdsaPrivate(inKey1Priv)
	util.AssertNoErr(t, err)
	t.Log("priv1", len(inKey1PrivDer), string(EncodeB64(inKey1PrivDer)))
	t.Log("pub1", len(inKey1PubDer), string(EncodeB64(inKey1PubDer)))

	// Generate pre-signature content
	var minBlock uint64 = 44
	outputs := []TxOut{
		{
			Value:         554,
			PublicKeyHash: DHash(outKey1PubDer),
		},
		{
			Value:         102,
			PublicKeyHash: DHash(outKey2PubDer),
		},
	}
	preSigHash := TxHashPreSig(minBlock, outputs)

	// Generate inputs with signatures
	sig1Asn, err := EcdsaSign(inKey1Priv, preSigHash)
	util.AssertNoErr(t, err)
	sig2Asn, err := EcdsaSign(inKey2Priv, preSigHash)
	util.AssertNoErr(t, err)
	t.Log("sig1", len(sig1Asn), string(EncodeB64(sig1Asn)))
	inputs := []TxIn{
		{
			OriginTxId:     DHash([]byte("Hello World")),
			OriginTxOutInd: 2,
			PublicKey:      inKey1PubDer,
			Signature:      sig1Asn,
		},
		{
			OriginTxId:     DHash([]byte("Hello World 123")),
			OriginTxOutInd: 3,
			PublicKey:      inKey2PubDer,
			Signature:      sig2Asn,
		},
	}

	// Generate final hash
	tx := Tx{
		MinBlock: minBlock,
		Inputs:   inputs,
		Outputs:  outputs,
	}
	txHash := tx.Hash()
	t.Log("txhash", len(txHash), HashHex(txHash), string(EncodeB64(txHash[:])))
}

// Test that a tx and components can be json serialized and deserialized.
func TestTxJson(t *testing.T) {
	originId, err := RandHash()
	util.AssertNoErr(t, err)
	outPkh1, err := RandHash()
	util.AssertNoErr(t, err)
	outPkh2, err := RandHash()
	util.AssertNoErr(t, err)
	tx := Tx{
		MinBlock: 443,
		Inputs: []TxIn{
			{
				OriginTxId:     originId,
				OriginTxOutInd: 2,
				PublicKey:      []byte("abc123"),
				Signature:      []byte("def456"),
				Value:          5223,
			},
			{
				OriginTxId:     originId,
				OriginTxOutInd: 3,
				PublicKey:      []byte("ghi789"),
				Signature:      []byte("jkl012"),
				Value:          3322,
			},
		},
		Outputs: []TxOut{
			{
				Value:         3500,
				PublicKeyHash: outPkh1,
			},
			{
				Value:         4500,
				PublicKeyHash: outPkh2,
			},
		},
	}
	txJs, err := json.Marshal(tx)
	util.AssertNoErr(t, err)
	txR := db.Tx{}
	err = json.Unmarshal(txJs, &txR)
	util.AssertNoErr(t, err)
	txRJs, err := json.Marshal(txR)
	util.AssertNoErr(t, err)
	util.Assert(t, bytes.Equal(txJs, txRJs), "serialization not preserved")
	t.Log(string(txJs))
}
