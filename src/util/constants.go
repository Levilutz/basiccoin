package util

import (
	"time"
)

// Types of the constants struct.
type ConstantsType struct {
	BlockReward          uint64        `json:"blockReward"`
	DebugLevel           int           `json:"debug"`
	HttpPort             int           `json:"httpPort"`
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
	BlockReward:          1000000,
	LocalAddr:            "localhost:21720",
	MinPeers:             3,
	MaxPeers:             8,
	MaxBlockVSize:        100000,
	MaxTxVSize:           5000,
	PeerPingFreq:         time.Second * 5,
	PrintPeersUpdateFreq: time.Second * 5,
	RuntimeID:            AssertUUID(),
	SeekNewPeersFreq:     time.Second * 10,
	Version:              "basiccoin:0.1.0",
}
