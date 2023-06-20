package peerfactory

import (
	"time"

	"github.com/levilutz/basiccoin/pkg/prot"
)

// Params to configure how we maintain our peer network.
type Params struct {
	ConnParams       prot.Params
	Listen           bool
	MinPeers         int
	MaxPeers         int
	SeekNewPeersFreq time.Duration
}

// Generate new production network params.
func ProdParams(listen bool) Params {
	return Params{
		ConnParams:       prot.NewParams(),
		Listen:           listen,
		MinPeers:         8,
		MaxPeers:         32,
		SeekNewPeersFreq: 15 * time.Second,
	}
}

// Generate new local dev network params.
func DevParams(listen bool) Params {
	return Params{
		ConnParams:       prot.NewParams(),
		Listen:           listen,
		MinPeers:         3,
		MaxPeers:         5,
		SeekNewPeersFreq: 5 * time.Second,
	}
}
