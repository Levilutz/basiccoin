package models

import (
	"encoding/json"

	"github.com/levilutz/basiccoin/pkg/core"
)

type BalanceResp struct {
	Balances map[core.HashT]uint64
}

type balanceRespJSON struct {
	Balances map[string]uint64 `json:"balances"`
}

func (r BalanceResp) MarshalJSON() ([]byte, error) {
	return json.Marshal(balanceRespJSON{
		Balances: core.MarshalHashTMap(r.Balances),
	})
}

func (r *BalanceResp) UnmarshalJSON(data []byte) error {
	raw := balanceRespJSON{}
	err := json.Unmarshal(data, &raw)
	if err != nil {
		return err
	}
	balances, err := core.UnmarshalHashTMap(raw.Balances)
	if err != nil {
		return err
	}
	r.Balances = balances
	return nil
}

type UtxosResp struct {
	Utxos map[core.Utxo]core.HashT
}

type utxosRespJSONItem struct {
	Utxo core.Utxo  `json:"utxo"`
	Pkh  core.HashT `json:"pkh"`
}

type utxosRespJSON struct {
	Utxos []utxosRespJSONItem `json:"utxos"`
}

func (r UtxosResp) MarshalJSON() ([]byte, error) {
	tuples := make([]utxosRespJSONItem, len(r.Utxos))
	i := 0
	for utxo, pkh := range r.Utxos {
		tuples[i] = utxosRespJSONItem{
			Utxo: utxo,
			Pkh:  pkh,
		}
		i++
	}
	return json.Marshal(utxosRespJSON{
		Utxos: tuples,
	})
}

func (r *UtxosResp) UnmarshalJSON(data []byte) error {
	raw := utxosRespJSON{}
	err := json.Unmarshal(data, &raw)
	if err != nil {
		return err
	}
	out := make(map[core.Utxo]core.HashT, len(raw.Utxos))
	for _, item := range raw.Utxos {
		out[item.Utxo] = item.Pkh
	}
	r.Utxos = out
	return nil
}

type TxConfirmsResp struct {
	Confirms map[core.HashT]uint64
}

type txConfirmsRespJSON struct {
	Confirms map[string]uint64 `json:"confirms"`
}

func (r TxConfirmsResp) MarshalJSON() ([]byte, error) {
	return json.Marshal(txConfirmsRespJSON{
		Confirms: core.MarshalHashTMap(r.Confirms),
	})
}

func (r *TxConfirmsResp) UnmarshalJSON(data []byte) error {
	raw := txConfirmsRespJSON{}
	err := json.Unmarshal(data, &raw)
	if err != nil {
		return err
	}
	confirms, err := core.UnmarshalHashTMap(raw.Confirms)
	if err != nil {
		return err
	}
	r.Confirms = confirms
	return nil
}

type TxIncludedBlockResp struct {
	IncludedBlocks map[core.HashT]core.HashT
}

type txIncludedBlockRespJSON struct {
	IncludedBlocks map[string]core.HashT `json:"includedBlocks"`
}

func (r TxIncludedBlockResp) MarshalJSON() ([]byte, error) {
	return json.Marshal(txIncludedBlockRespJSON{
		IncludedBlocks: core.MarshalHashTMap(r.IncludedBlocks),
	})
}

func (r *TxIncludedBlockResp) UnmarshalJSON(data []byte) error {
	raw := txIncludedBlockRespJSON{}
	err := json.Unmarshal(data, &raw)
	if err != nil {
		return err
	}
	includedBlocks, err := core.UnmarshalHashTMap(raw.IncludedBlocks)
	if err != nil {
		return err
	}
	r.IncludedBlocks = includedBlocks
	return nil
}

type GetTxResp struct {
	Txs map[core.HashT]core.Tx
}

type getTxRespJSON struct {
	Txs map[string]core.Tx `json:"txs"`
}

func (r GetTxResp) MarshalJSON() ([]byte, error) {
	return json.Marshal(getTxRespJSON{
		Txs: core.MarshalHashTMap(r.Txs),
	})
}

func (r *GetTxResp) UnmarshalJSON(data []byte) error {
	raw := getTxRespJSON{}
	err := json.Unmarshal(data, &raw)
	if err != nil {
		return err
	}
	txs, err := core.UnmarshalHashTMap(raw.Txs)
	if err != nil {
		return err
	}
	r.Txs = txs
	return nil
}

type GetMerkleResp struct {
	Merkles map[core.HashT]core.MerkleNode
}

type getMerkleRespJSON struct {
	Merkles map[string]core.MerkleNode `json:"merkles"`
}

func (r GetMerkleResp) MarshalJSON() ([]byte, error) {
	return json.Marshal(getMerkleRespJSON{
		Merkles: core.MarshalHashTMap(r.Merkles),
	})
}

func (r *GetMerkleResp) UnmarshalJSON(data []byte) error {
	raw := getMerkleRespJSON{}
	err := json.Unmarshal(data, &raw)
	if err != nil {
		return err
	}
	merkles, err := core.UnmarshalHashTMap(raw.Merkles)
	if err != nil {
		return err
	}
	r.Merkles = merkles
	return nil
}

type GetBlockResp struct {
	Blocks map[core.HashT]core.Block
}

type getBlockRespJSON struct {
	Blocks map[string]core.Block `json:"blocks"`
}

func (r GetBlockResp) MarshalJSON() ([]byte, error) {
	return json.Marshal(getBlockRespJSON{
		Blocks: core.MarshalHashTMap(r.Blocks),
	})
}

func (r *GetBlockResp) UnmarshalJSON(data []byte) error {
	raw := getBlockRespJSON{}
	err := json.Unmarshal(data, &raw)
	if err != nil {
		return err
	}
	blocks, err := core.UnmarshalHashTMap(raw.Blocks)
	if err != nil {
		return err
	}
	r.Blocks = blocks
	return nil
}

type RichListResp struct {
	RichList map[core.HashT]uint64
}

type richListRespJSON struct {
	RichList map[string]uint64 `json:"richList"`
}

func (r RichListResp) MarshalJSON() ([]byte, error) {
	return json.Marshal(richListRespJSON{
		RichList: core.MarshalHashTMap(r.RichList),
	})
}

func (r *RichListResp) UnmarshalJSON(data []byte) error {
	raw := richListRespJSON{}
	err := json.Unmarshal(data, &raw)
	if err != nil {
		return err
	}
	richList, err := core.UnmarshalHashTMap(raw.RichList)
	if err != nil {
		return err
	}
	r.RichList = richList
	return nil
}
