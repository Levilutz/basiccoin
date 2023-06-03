package peer

import (
	"github.com/levilutz/basiccoin/src/util"
)

type PeerMessage interface {
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
	addr := ""
	if util.Constants.Listen {
		addr = util.Constants.LocalAddr
	}
	return HelloMessage{
		RuntimeID: util.Constants.RuntimeID,
		Version:   util.Constants.Version,
		Addr:      addr,
	}
}

// Receive a HelloMessage from the channel
func ReceiveHelloMessage(pc *PeerConn) (HelloMessage, error) {
	return HelloMessage{
		RuntimeID: pc.RetryReadStringLine(7),
		Version:   pc.RetryReadStringLine(7),
		Addr:      pc.RetryReadStringLine(7),
	}, pc.Err()
}

// Transmit a HelloMessage over the channel
func (msg HelloMessage) Transmit(pc *PeerConn) error {
	pc.TransmitStringLine(msg.RuntimeID)
	pc.TransmitStringLine(msg.Version)
	pc.TransmitStringLine(msg.Addr)
	return pc.Err()
}
