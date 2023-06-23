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
