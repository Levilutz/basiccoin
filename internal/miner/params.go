package miner

import "github.com/levilutz/basiccoin/pkg/core"

// Params to configure the miner.
type Params struct {
	// Public key hash to pay out block rewards to.
	PayoutPkh core.HashT
}

// Generate params.
func NewParams(payoutPkh core.HashT) Params {
	return Params{
		PayoutPkh: payoutPkh,
	}
}
