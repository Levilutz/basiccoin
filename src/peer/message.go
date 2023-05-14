package peer

import (
	"github.com/levilutz/basiccoin/src/util"
)

type PeerMessage interface {
	GetName() string
	Transmit(pc *PeerConn) error
}

// Receive base64(json(message)) from a single line
func receiveStandardMessage[R PeerMessage](pc *PeerConn) (R, error) {
	// Cannot be method until golang allows type params on methods
	var content R
	data := pc.RetryReadLine(7)
	if err := pc.Err(); err != nil {
		return content, err
	}
	return util.UnJsonB64[R](data)
}

// HelloMessage

type HelloPeerMessage struct {
	RuntimeID string `json:"runtimeID"`
	Version   string `json:"version"`
	Addr      string `json:"addr"`
}

// Construct a HelloMessage
func NewHelloMessage() HelloPeerMessage {
	return HelloPeerMessage{
		RuntimeID: util.Constants.RuntimeID,
		Version:   util.Constants.Version,
		Addr:      util.Constants.LocalAddr,
	}
}

// Receive a HelloMessage from the channel
func ReceiveHelloMessage(pc *PeerConn) (HelloPeerMessage, error) {
	return receiveStandardMessage[HelloPeerMessage](pc)
}

// Get the HelloMessage's name
func (msg HelloPeerMessage) GetName() string {
	return "hello"
}

// Transmit a HelloMessage over the channel, including name
func (msg HelloPeerMessage) Transmit(pc *PeerConn) error {
	pc.TransmitStringLine("cmd:" + msg.GetName())
	data, err := util.JsonB64(msg)
	if err != nil {
		return err
	}
	pc.TransmitLine(data)
	return pc.Err()
}
