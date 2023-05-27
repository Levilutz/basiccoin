package util

import (
	"math/big"
	"time"
)

// Types of the constants struct
type ConstantsType struct {
	DebugNetwork         bool          `json:"debugNetwork"`
	LocalAddr            string        `json:"localAddr"`
	Listen               bool          `json:"listen"`
	MinPeers             int           `json:"maxPeers"`
	MaxPeers             int           `json:"minPeers"`
	MaxBlockTxs          uint64        `json:"maxBlockTxs"`
	MaxBlockVSize        uint64        `json:"maxBlockVSize"`
	MaxTreeSize          uint64        `json:"maxTreeSize"`
	MaxTxVSize           uint64        `json:"maxTxVSize"`
	PeerPingFreq         time.Duration `json:"peerPingFreq"`
	PrintPeersUpdateFreq time.Duration `json:"printPeersUpdateFreq"`
	RuntimeID            string        `json:"runtimeID"`
	SeedAddr             string        `json:"seedAddr"`
	SeekNewPeersFreq     time.Duration `json:"seekNewPeersFreq"`
	Version              string        `json:"version"`
}

// Program-wide constants, should be set on startup
var Constants = ConstantsType{
	DebugNetwork:         false,
	LocalAddr:            "localhost:21720",
	MinPeers:             3,
	MaxPeers:             8,
	MaxBlockTxs:          256,
	MaxBlockVSize:        100000,
	MaxTxVSize:           5000,
	Listen:               true,
	PeerPingFreq:         time.Second * 5,
	PrintPeersUpdateFreq: time.Second * 5,
	RuntimeID:            AssertUUID(),
	SeedAddr:             "",
	SeekNewPeersFreq:     time.Second * 10,
	Version:              "0.1.0",
}

func InitComputedConstants() {
	Constants.MaxTreeSize = 2*Constants.MaxBlockTxs - 1
}

func BigInt2_256() *big.Int {
	out := &big.Int{}
	out.SetString(
		"ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
		16,
	)
	return out
}
