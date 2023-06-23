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
