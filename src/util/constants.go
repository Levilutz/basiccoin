package util

import "time"

// Types of the constants struct
type ConstantsType struct {
	DebugManagerLoop     bool          `json:"debugManagerLoop"`
	DebugNetwork         bool          `json:"debugNetwork"`
	DebugPeerLoop        bool          `json:"debugPeerLoop"`
	DebugTicker          bool          `json:"debugTicker"`
	FilterKnownPeersFreq time.Duration `json:"filterKnownPeersFreq"`
	LocalAddr            string        `json:"localAddr"`
	Listen               bool          `json:"listen"`
	PeerBusBufferSize    int           `json:"peerBusBufferSize"`
	PeerListenFreq       time.Duration `json:"peerListenFreq"`
	PeerPingFreq         time.Duration `json:"peerPingFreq"`
	PrintPeersUpdateFreq time.Duration `json:"printPeersUpdateFreq"`
	RuntimeID            string        `json:"runtimeID"`
	SeedAddr             string        `json:"seedAddr"`
	Version              string        `json:"version"`
}

// Program-wide constants, should be set on startup
var Constants = ConstantsType{
	DebugManagerLoop:     false,
	DebugNetwork:         false,
	DebugPeerLoop:        false,
	DebugTicker:          false,
	FilterKnownPeersFreq: time.Second * 10,
	LocalAddr:            "localhost:21720",
	Listen:               true,
	PeerBusBufferSize:    100,
	PeerListenFreq:       time.Millisecond * 100,
	PeerPingFreq:         time.Second * 5,
	PrintPeersUpdateFreq: time.Second * 5,
	RuntimeID:            AssertUUID(),
	SeedAddr:             "",
	Version:              "0.1.0",
}
