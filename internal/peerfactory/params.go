package peerfactory

import (
	"fmt"
	"time"

	"github.com/levilutz/basiccoin/pkg/core"
)

// Params to configure how we maintain our peer network.
type Params struct {
	// Whether to setup the debug flag on all new connections.
	DebugConns bool

	// Whether to listen for inbound connections
	Listen bool

	// The local address to broadcast.
	// If listen is true and this is empty, it's discovered from our first peer.
	LocalAddr string

	// Below this number of peers, actively seek new ones.
	MinPeers int

	// At or above this number of peers, reject inbound connections.
	MaxPeers int

	// An id to uniquely identify this node.
	RuntimeId string

	// The frequency with which we seek new peers, if appropriate.
	SeekNewPeersFreq time.Duration
}

// Generate a random runtime id.
func NewRuntimeId() string {
	runtimeId := core.NewHashTRand().String()
	fmt.Println("RuntimeId:", runtimeId)
	return runtimeId
}

// Generate new production network params.
func ProdParams(listen bool, localAddr string) Params {
	return Params{
		DebugConns:       false,
		Listen:           listen,
		LocalAddr:        localAddr,
		MinPeers:         8,
		MaxPeers:         32,
		RuntimeId:        NewRuntimeId(),
		SeekNewPeersFreq: 15 * time.Second,
	}
}

// Generate new local dev network params.
func DevParams(listen bool, localAddr string) Params {
	return Params{
		DebugConns:       false,
		Listen:           listen,
		LocalAddr:        localAddr,
		MinPeers:         3,
		MaxPeers:         5,
		RuntimeId:        NewRuntimeId(),
		SeekNewPeersFreq: 5 * time.Second,
	}
}
