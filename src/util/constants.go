package util

import "time"

// Types of the constants struct
type ConstantsType struct {
	LocalAddr         string        `json:"localAddr"`
	PeerBusBufferSize int           `json:"peerBusBufferSize"`
	PeerListenFreq    time.Duration `json:"peerListenFreq"`
	PeerPingFreq      time.Duration `json:"peerPingFreq"`
	RuntimeID         string        `json:"runtimeID"`
	Version           string        `json:"version"`
}

// Program-wide constants, should be set on startup
var Constants = ConstantsType{
	LocalAddr:         "localhost:21720",
	PeerBusBufferSize: 100,
	PeerListenFreq:    time.Millisecond * time.Duration(100),
	PeerPingFreq:      time.Second * time.Duration(5),
	RuntimeID:         AssertUUID(),
	Version:           "0.1.0",
}
