package util

import (
	"time"
)

// Types of the constants struct.
type ConstantsType struct {
	DebugLevel           int           `json:"debug"`
	HttpPort             int           `json:"httpPort"`
	LocalAddr            string        `json:"localAddr"`
	Listen               bool          `json:"listen"`
	Miners               int           `json:"miners"`
	MinPeers             int           `json:"maxPeers"`
	MaxPeers             int           `json:"minPeers"`
	PeerPingFreq         time.Duration `json:"peerPingFreq"`
	PrintPeersUpdateFreq time.Duration `json:"printPeersUpdateFreq"`
	RuntimeID            string        `json:"runtimeID"`
	SeedAddr             string        `json:"seedAddr"`
	SeekNewPeersFreq     time.Duration `json:"seekNewPeersFreq"`
	Version              string        `json:"version"`
}

// Program-wide constants set by user.
var Constants = ConstantsType{
	LocalAddr:            "localhost:21720",
	MinPeers:             3,
	MaxPeers:             8,
	PeerPingFreq:         time.Second * 5,
	PrintPeersUpdateFreq: time.Second * 5,
	RuntimeID:            AssertUUID(),
	SeekNewPeersFreq:     time.Second * 10,
	Version:              "basiccoin:0.1.0",
}
