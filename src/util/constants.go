package util

import (
	"math/big"
	"time"
)

// Types of the constants struct.
type ConstantsType struct {
	BlockReward          uint64        `json:"blockReward"`
	Debug                bool          `json:"debug"`
	LocalAddr            string        `json:"localAddr"`
	Listen               bool          `json:"listen"`
	Miners               int           `json:"miners"`
	MinPeers             int           `json:"maxPeers"`
	MaxPeers             int           `json:"minPeers"`
	MaxBlockVSize        uint64        `json:"maxBlockVSize"`
	MaxTxVSize           uint64        `json:"maxTxVSize"`
	PeerPingFreq         time.Duration `json:"peerPingFreq"`
	PrintPeersUpdateFreq time.Duration `json:"printPeersUpdateFreq"`
	RuntimeID            string        `json:"runtimeID"`
	SeedAddr             string        `json:"seedAddr"`
	SeekNewPeersFreq     time.Duration `json:"seekNewPeersFreq"`
	Version              string        `json:"version"`
}

// Program-wide constants set by user.
var Constants = ConstantsType{
	BlockReward:          1028,
	Debug:                false,
	LocalAddr:            "localhost:21720",
	MinPeers:             3,
	MaxPeers:             8,
	MaxBlockVSize:        100000,
	MaxTxVSize:           5000,
	PeerPingFreq:         time.Second * 5,
	PrintPeersUpdateFreq: time.Second * 5,
	RuntimeID:            AssertUUID(),
	SeedAddr:             "",
	SeekNewPeersFreq:     time.Second * 10,
	Version:              "0.1.0",
}

// Compute 2^256 as a big.Int.
func BigInt2_256() *big.Int {
	out := &big.Int{}
	out.SetString(
		"ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
		16,
	)
	return out
}
