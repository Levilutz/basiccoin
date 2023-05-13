package main

import (
	"github.com/levilutz/basiccoin/src/util"
)

// HelloMessage

var HelloMessageName = "hello"

type HelloMessage struct {
	RuntimeID string `json:"runtimeID"`
	Version   string `json:"version"`
	Addr      string `json:"addr"`
}

// Construct a HelloMessage
func NewHelloMessage() HelloMessage {
	return HelloMessage{
		RuntimeID: util.Constants.RuntimeID,
		Version:   util.Constants.Version,
		Addr:      util.Constants.LocalAddr,
	}
}

// Receive a HelloMessage from the channel
func ReceiveHelloMessage(pc PeerConn) (HelloMessage, error) {
	return PeerConnReceiveStandardMessage[HelloMessage](pc, HelloMessageName)
}

// Transmit a HelloMessage over the channel, including name
func (msg HelloMessage) Transmit(pc PeerConn) error {
	return pc.TransmitStandardMessage(HelloMessageName, msg)
}
