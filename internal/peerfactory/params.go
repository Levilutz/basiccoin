package peerfactory

import (
	"time"
)

// Params to configure how we maintain our peer network.
type Params struct {
	Listen           bool
	MinPeers         int
	MaxPeers         int
	RuntimeId        string
	SeekNewPeersFreq time.Duration
}

// Generate new production network params.
func ProdParams(listen bool, runtimeId string) Params {
	return Params{
		Listen:           listen,
		MinPeers:         8,
		MaxPeers:         32,
		RuntimeId:        runtimeId,
		SeekNewPeersFreq: 15 * time.Second,
	}
}

// Generate new local dev network params.
func DevParams(listen bool, runtimeId string) Params {
	return Params{
		Listen:           listen,
		MinPeers:         3,
		MaxPeers:         5,
		RuntimeId:        runtimeId,
		SeekNewPeersFreq: 5 * time.Second,
	}
}
