package util

import "time"

// Types of the constants struct
type ConstantsType struct {
	DebugNetwork         bool          `json:"debugNetwork"`
	LocalAddr            string        `json:"localAddr"`
	Listen               bool          `json:"listen"`
	MinPeers             int           `json:"maxPeers"`
	MaxPeers             int           `json:"minPeers"`
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
	Listen:               true,
	PeerPingFreq:         time.Second * 5,
	PrintPeersUpdateFreq: time.Second * 5,
	RuntimeID:            AssertUUID(),
	SeedAddr:             "",
	SeekNewPeersFreq:     time.Second * 10,
	Version:              "0.1.0",
}
