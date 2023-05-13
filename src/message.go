package main

import (
	"fmt"

	"github.com/levilutz/basiccoin/src/util"
)

// Generic helpers

// Consume the next line and assert that it matches msg
func ConsumeExpected(pc util.PeerConn, msg string) error {
	data, err := util.RetryReadLine(pc, 8)
	if err != nil {
		return err
	}
	if string(data) != msg {
		return fmt.Errorf(
			"expected '%s', received '%s'", msg, string(data),
		)
	}
	return nil
}

// Transmit a simple string
func TransmitSimpleMessage(pc util.PeerConn, msg string) error {
	content := []byte(msg + "\n")
	_, err := pc.W.Write(content)
	if err != nil {
		return err
	}
	return pc.W.Flush()
}

// Receive base64(json(message)) from a single line
func ReceiveStandardMessage[R any](pc util.PeerConn, msgName string) (R, error) {
	var content R
	data, err := util.RetryReadLine(pc, 8)
	if err != nil {
		return content, err
	}
	return util.UnJsonB64[R](data)
}

// Transmit msgName then base64(json(message)) in a single line each
func TransmitStandardMessage(pc util.PeerConn, msgName string, msg any) error {
	data, err := util.JsonB64(msg)
	if err != nil {
		return err
	}
	content := []byte(msgName + "\n")
	content = append(content, data...)
	content = append(content, byte('\n'))
	_, err = pc.W.Write(content)
	if err != nil {
		return err
	}
	return pc.W.Flush()
}

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
func ReceiveHelloMessage(pc util.PeerConn) (HelloMessage, error) {
	return ReceiveStandardMessage[HelloMessage](pc, HelloMessageName)
}

// Transmit a HelloMessage over the channel, including name
func (msg HelloMessage) Transmit(pc util.PeerConn) error {
	return TransmitStandardMessage(pc, HelloMessageName, msg)
}
