package inv

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"github.com/levilutz/basiccoin/pkg/core"
)

func BlockRecordFromString(raw string) (record BlockRecord, err error) {
	rows := strings.Split(strings.Trim(raw, "\n"), "\n")
	if len(rows) != 8 {
		return BlockRecord{}, fmt.Errorf("incorrect number of rows: %d", len(rows))
	}
	prevBlockId, err := core.NewHashTFromString(rows[0])
	if err != nil {
		return BlockRecord{}, fmt.Errorf("failed to parse PrevBlockId: %s", err)
	}
	merkleRoot, err := core.NewHashTFromString(rows[1])
	if err != nil {
		return BlockRecord{}, fmt.Errorf("failed to parse MerkleRoot: %s", err)
	}
	target, err := core.NewHashTFromString(rows[2])
	if err != nil {
		return BlockRecord{}, fmt.Errorf("failed to parse Target: %s", err)
	}
	noise, err := core.NewHashTFromString(rows[3])
	if err != nil {
		return BlockRecord{}, fmt.Errorf("failed to parse Noise: %s", err)
	}
	nonce, err := strconv.ParseUint(rows[4], 10, 64)
	if err != nil {
		return BlockRecord{}, fmt.Errorf("failed to parse Nonce: %s", err)
	}
	minedTime, err := strconv.ParseUint(rows[5], 10, 64)
	if err != nil {
		return BlockRecord{}, fmt.Errorf("failed to parse MinedTime: %s", err)
	}
	height, err := strconv.ParseUint(rows[6], 10, 64)
	if err != nil {
		return BlockRecord{}, fmt.Errorf("failed to parse Height: %s", err)
	}
	totalWork, err := core.NewHashTFromString(rows[7])
	if err != nil {
		return BlockRecord{}, fmt.Errorf("failed to parse TotalWork: %s", err)
	}
	return BlockRecord{
		Block: core.Block{
			PrevBlockId: prevBlockId,
			MerkleRoot:  merkleRoot,
			Target:      target,
			Noise:       noise,
			Nonce:       nonce,
			MinedTime:   minedTime,
		},
		Height:    height,
		TotalWork: totalWork,
	}, nil
}

func (b BlockRecord) String() string {
	return strings.Join([]string{
		b.Block.PrevBlockId.String(),
		b.Block.MerkleRoot.String(),
		b.Block.Target.String(),
		b.Block.Noise.String(),
		strconv.FormatUint(b.Block.Nonce, 10),
		strconv.FormatUint(b.Block.MinedTime, 10),
		strconv.FormatUint(b.Height, 10),
		b.TotalWork.String(),
	}, "\n")
}

func MerkleRecordFromString(raw string) (record MerkleRecord, err error) {
	rows := strings.Split(strings.Trim(raw, "\n"), "\n")
	if len(rows) != 3 {
		return MerkleRecord{}, fmt.Errorf("incorrect number of rows: %d", len(rows))
	}
	lChild, err := core.NewHashTFromString(rows[0])
	if err != nil {
		return MerkleRecord{}, fmt.Errorf("failed to parse LChild: %s", err)
	}
	rChild, err := core.NewHashTFromString(rows[1])
	if err != nil {
		return MerkleRecord{}, fmt.Errorf("failed to parse RChild: %s", err)
	}
	vSize, err := strconv.ParseUint(rows[2], 10, 64)
	if err != nil {
		return MerkleRecord{}, fmt.Errorf("failed to parse VSize: %s", err)
	}
	return MerkleRecord{
		Merkle: core.MerkleNode{
			LChild: lChild,
			RChild: rChild,
		},
		VSize: vSize,
	}, nil
}

func (m MerkleRecord) String() string {
	return strings.Join([]string{
		m.Merkle.LChild.String(),
		m.Merkle.RChild.String(),
		strconv.FormatUint(m.VSize, 10),
	}, "\n")
}

func TxRecordFromString(raw string) (record TxRecord, err error) {
	rows := strings.Split(strings.Trim(raw, "\n"), "\n")
	if len(rows) < 5 {
		return TxRecord{}, fmt.Errorf("too few rows: %d", len(rows))
	}
	vSize, err := strconv.ParseUint(rows[0], 10, 64)
	if err != nil {
		return TxRecord{}, fmt.Errorf("failed to parse VSize: %s", err)
	}
	var isCoinbase bool
	if rows[1] == "true" {
		isCoinbase = true
	} else if rows[1] == "false" {
		isCoinbase = false
	} else {
		return TxRecord{}, fmt.Errorf("failed to parse IsCoinbase")
	}
	minBlock, err := strconv.ParseUint(rows[2], 10, 64)
	if err != nil {
		return TxRecord{}, fmt.Errorf("failed to parse MinBlock: %s", err)
	}
	numInputs, err := strconv.Atoi(rows[3])
	if err != nil {
		return TxRecord{}, fmt.Errorf("failed to parse NumInputs: %s", err)
	}
	numOutputs, err := strconv.Atoi(rows[4])
	if err != nil {
		return TxRecord{}, fmt.Errorf("failed to parse NumOutputs: %s", err)
	}
	expectRows := 5 + numInputs*5 + numOutputs*2
	if len(rows) != expectRows {
		return TxRecord{}, fmt.Errorf("expected %d rows, got %d", expectRows, len(rows))
	}
	inputs := make([]core.TxIn, numInputs)
	outputs := make([]core.TxOut, numOutputs)
	currentRow := 5
	for i := range inputs {
		txId, err := core.NewHashTFromString(rows[currentRow])
		currentRow++
		if err != nil {
			return TxRecord{}, fmt.Errorf("failed to parse input %d TxId: %s", i, err)
		}
		ind, err := strconv.ParseUint(rows[currentRow], 10, 64)
		currentRow++
		if err != nil {
			return TxRecord{}, fmt.Errorf("failed to parse input %d Ind: %s", i, err)
		}
		value, err := strconv.ParseUint(rows[currentRow], 10, 64)
		currentRow++
		if err != nil {
			return TxRecord{}, fmt.Errorf("failed to parse input %d Value: %s", i, err)
		}
		publicKey, err := base64.StdEncoding.DecodeString(rows[currentRow])
		currentRow++
		if err != nil {
			return TxRecord{}, fmt.Errorf("failed to parse input %d PublicKey: %s", i, err)
		}
		signature, err := base64.StdEncoding.DecodeString(rows[currentRow])
		currentRow++
		if err != nil {
			return TxRecord{}, fmt.Errorf("failed to parse input %d Signature: %s", i, err)
		}
		inputs[i] = core.TxIn{
			Utxo: core.Utxo{
				TxId:  txId,
				Ind:   ind,
				Value: value,
			},
			PublicKey: publicKey,
			Signature: signature,
		}
	}
	for i := range outputs {
		value, err := strconv.ParseUint(rows[currentRow], 10, 64)
		currentRow++
		if err != nil {
			return TxRecord{}, fmt.Errorf("failed to parse output %d Value: %s", i, err)
		}
		pkh, err := core.NewHashTFromString(rows[currentRow])
		currentRow++
		if err != nil {
			return TxRecord{}, fmt.Errorf("failed to parse input %d PublicKeyHash: %s", i, err)
		}
		outputs[i] = core.TxOut{
			Value:         value,
			PublicKeyHash: pkh,
		}
	}
	return TxRecord{
		Tx: core.Tx{
			IsCoinbase: isCoinbase,
			MinBlock:   minBlock,
			Inputs:     inputs,
			Outputs:    outputs,
		},
		VSize: vSize,
	}, nil
}

func (t TxRecord) String() string {
	rows := make([]string, 5)
	rows[0] = strconv.FormatUint(t.VSize, 10)
	rows[1] = fmt.Sprintf("%t", t.Tx.IsCoinbase)
	rows[2] = strconv.FormatUint(t.Tx.MinBlock, 10)
	rows[3] = strconv.Itoa(len(t.Tx.Inputs))
	rows[4] = strconv.Itoa(len(t.Tx.Outputs))
	for _, input := range t.Tx.Inputs {
		rows = append(rows, []string{
			input.Utxo.TxId.String(),
			strconv.FormatUint(input.Utxo.Ind, 10),
			strconv.FormatUint(input.Utxo.Value, 10),
			base64.StdEncoding.EncodeToString(input.PublicKey),
			base64.StdEncoding.EncodeToString(input.Signature),
		}...)
	}
	for _, output := range t.Tx.Outputs {
		rows = append(rows, []string{
			strconv.FormatUint(output.Value, 10),
			output.PublicKeyHash.String(),
		}...)
	}
	return strings.Join(rows, "\n")
}
