package main

import (
	"github.com/levilutz/basiccoin/src/util"
)

type Message interface {
	GetName() string
	Transmit(pc *PeerConn) error
}

// HelloMessage

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
func ReceiveHelloMessage(pc *PeerConn) (HelloMessage, error) {
	helloMsg := PeerConnReceiveStandardMessage[HelloMessage](pc)
	return helloMsg, pc.Err()
}

// Get the HelloMessage's name
func (msg HelloMessage) GetName() string {
	return "hello"
}

// Transmit a HelloMessage over the channel, including name
func (msg HelloMessage) Transmit(pc *PeerConn) error {
	pc.TransmitStandardMessage(msg)
	return pc.Err()
}
