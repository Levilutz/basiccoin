package inv_test

import (
	"bytes"
	"testing"
	"time"

	. "github.com/levilutz/basiccoin/internal/inv"
	"github.com/levilutz/basiccoin/pkg/core"
	"github.com/levilutz/basiccoin/pkg/util"
)

func TestSerializeBlockRecord(t *testing.T) {
	record := BlockRecord{
		Block: core.Block{
			PrevBlockId: core.NewHashTRand(),
			MerkleRoot:  core.NewHashTRand(),
			Target:      core.NewHashTRand(),
			Noise:       core.NewHashTRand(),
			Nonce:       51029519285,
			MinedTime:   uint64(time.Now().Unix()),
		},
		Height:    10502,
		TotalWork: core.NewHashTRand(),
	}
	ser := record.String()
	recon, err := BlockRecordFromString(ser)
	util.Assert(t, err == nil, "failed to reconstruct: %s", err)
	util.Assert(t, recon.Block.PrevBlockId.Eq(record.Block.PrevBlockId), "PrevBlockId mismatch")
	util.Assert(t, recon.Block.MerkleRoot.Eq(record.Block.MerkleRoot), "MerkleRoot mismatch")
	util.Assert(t, recon.Block.Target.Eq(record.Block.Target), "Target mismatch")
	util.Assert(t, recon.Block.Noise.Eq(record.Block.Noise), "Noise mismatch")
	util.Assert(t, recon.Block.Nonce == record.Block.Nonce, "Nonce mismatch")
	util.Assert(t, recon.Block.MinedTime == record.Block.MinedTime, "MinedTime mismatch")
	util.Assert(t, recon.Height == record.Height, "Height mismatch")
	util.Assert(t, recon.TotalWork.Eq(record.TotalWork), "PrevBlockId mismatch")
	util.Assert(t, recon.Block.Hash().Eq(record.Block.Hash()), "Hash mismatch")
}

func TestSerializeMerkleRecord(t *testing.T) {
	record := MerkleRecord{
		Merkle: core.MerkleNode{
			LChild: core.NewHashTRand(),
			RChild: core.NewHashTRand(),
		},
		VSize: 85712895,
	}
	ser := record.String()
	recon, err := MerkleRecordFromString(ser)
	util.Assert(t, err == nil, "failed to reconstruct: %s", err)
	util.Assert(t, recon.Merkle.LChild.Eq(record.Merkle.LChild), "LChild mismatch")
	util.Assert(t, recon.Merkle.RChild.Eq(record.Merkle.RChild), "RChild mismatch")
	util.Assert(t, recon.VSize == record.VSize, "VSize mismatch")
	util.Assert(t, recon.Merkle.Hash().Eq(record.Merkle.Hash()), "Hash mismatch")
}

func TestSerializeTxRecord(t *testing.T) {
	tx := core.Tx{
		IsCoinbase: false,
		MinBlock:   4124,
		Inputs: []core.TxIn{
			{
				Utxo: core.Utxo{
					TxId:  core.NewHashTRand(),
					Ind:   5,
					Value: 500,
				},
				PublicKey: []byte("pubKey1"),
				Signature: []byte("sig1"),
			},
			{
				Utxo: core.Utxo{
					TxId:  core.NewHashTRand(),
					Ind:   7,
					Value: 550,
				},
				PublicKey: []byte("pubKey2"),
				Signature: []byte("sig2"),
			},
		},
		Outputs: []core.TxOut{
			{
				Value:         400,
				PublicKeyHash: core.NewHashTRand(),
			},
			{
				Value:         450,
				PublicKeyHash: core.NewHashTRand(),
			},
		},
	}
	record := TxRecord{
		Tx:    tx,
		VSize: tx.VSize(),
	}
	ser := record.String()
	recon, err := TxRecordFromString(ser)
	util.Assert(t, err == nil, "failed to reconstruct: %s", err)
	util.Assert(t, recon.Tx.IsCoinbase == record.Tx.IsCoinbase, "IsCoinbase mismatch")
	util.Assert(t, recon.Tx.MinBlock == record.Tx.MinBlock, "MinBlock mismatch")
	util.Assert(t, len(recon.Tx.Inputs) == len(record.Tx.Inputs), "NumInputs mismatch")
	util.Assert(t, len(recon.Tx.Outputs) == len(record.Tx.Outputs), "NumOutputs mismatch")
	for i := range recon.Tx.Inputs {
		util.Assert(t, recon.Tx.Inputs[i].Utxo.TxId.Eq(record.Tx.Inputs[i].Utxo.TxId), "Input %d TxId mismatch", i)
		util.Assert(t, recon.Tx.Inputs[i].Utxo.Ind == record.Tx.Inputs[i].Utxo.Ind, "Input %d Ind mismatch", i)
		util.Assert(t, recon.Tx.Inputs[i].Utxo.Value == record.Tx.Inputs[i].Utxo.Value, "Input %d Value mismatch", i)
		util.Assert(t, bytes.Equal(recon.Tx.Inputs[i].PublicKey, record.Tx.Inputs[i].PublicKey), "Input %d PublicKey mismatch", i)
		util.Assert(t, bytes.Equal(recon.Tx.Inputs[i].Signature, record.Tx.Inputs[i].Signature), "Input %d Signature mismatch", i)
	}
	for i := range recon.Tx.Outputs {
		util.Assert(t, recon.Tx.Outputs[i].Value == record.Tx.Outputs[i].Value, "Output %d Value mismatch", i)
		util.Assert(t, recon.Tx.Outputs[i].PublicKeyHash.Eq(record.Tx.Outputs[i].PublicKeyHash), "Output %d PublicKeyHash mismatch", i)
	}
	util.Assert(t, recon.Tx.Hash().Eq(record.Tx.Hash()), "Hash mismatch")
}
