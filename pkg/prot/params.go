package prot

import "github.com/levilutz/basiccoin/pkg/core"

type Params struct {
	RuntimeID string `json:"runtimeId"` // An id to uniquely identify this node.
}

func NewParams() Params {
	return Params{
		RuntimeID: core.NewHashTRand().String(),
	}
}
