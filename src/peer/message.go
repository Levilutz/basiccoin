package peer

import (
	"github.com/levilutz/basiccoin/src/util"
)

type PeerMessage interface {
	Transmit(pc *PeerConn) error
}

// Receive base64(json(message)) from a single line
func receiveStandardMessage[R PeerMessage](pc *PeerConn) (R, error) {
	// Cannot be method until golang allows type params on methods
	var content R
	data := pc.RetryReadLine(7)
	if pc.HasErr() {
		return content, pc.Err()
	}
	return util.UnJsonB64[R](data)
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
	return receiveStandardMessage[HelloMessage](pc)
}

// Transmit a HelloMessage over the channel
func (msg HelloMessage) Transmit(pc *PeerConn) error {
	data, err := util.JsonB64(msg)
	if err != nil {
		return err
	}
	pc.TransmitLine(data)
	return pc.Err()
}

// AddrsMessage

type AddrsMessage struct {
	PeerAddrs []string `json:"peerAddrs"`
}

// Construct an AddrsMessage
func ReceiveAddrsMessage(pc *PeerConn) (AddrsMessage, error) {
	numAddrs := pc.RetryReadIntLine(7)
	if pc.HasErr() {
		return AddrsMessage{}, pc.Err()
	}
	addrs := make([]string, numAddrs)
	for i := 0; i < numAddrs; i++ {
		addrs[i] = pc.RetryReadStringLine(7)
	}
	pc.ConsumeExpected("fin:addrs")
	return AddrsMessage{
		PeerAddrs: addrs,
	}, pc.Err()
}

// Transmit an AddrsMessage over the channel
func (msg AddrsMessage) Transmit(pc *PeerConn) error {
	pc.TransmitIntLine(len(msg.PeerAddrs))
	for _, addr := range msg.PeerAddrs {
		if addr == "fin:addrs" {
			continue
		}
		pc.TransmitStringLine(addr)
	}
	pc.TransmitStringLine("fin:addrs")
	return pc.Err()
}
