package main

import (
	"bufio"
	"fmt"

	"github.com/levilutz/basiccoin/src/util"
)

// Generic helpers

// Consume the next line and assert that it matches msg
func ConsumeExpected(r *bufio.Reader, msg string) error {
	data, err := util.RetryReadLine(r, 8)
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
func TransmitSimpleMessage(w *bufio.Writer, msg string) error {
	content := []byte(msg + "\n")
	_, err := w.Write(content)
	if err != nil {
		return err
	}
	return w.Flush()
}

// Receive base64(json(message)) from a single line
func ReceiveStandardMessage[R any](r *bufio.Reader, msgName string) (R, error) {
	var content R
	data, err := util.RetryReadLine(r, 8)
	if err != nil {
		return content, err
	}
	return util.UnJsonB64[R](data)
}

// Transmit msgName then base64(json(message)) in a single line each
func TransmitStandardMessage(w *bufio.Writer, msgName string, msg any) error {
	data, err := util.JsonB64(msg)
	if err != nil {
		return err
	}
	content := []byte(msgName + "\n")
	content = append(content, data...)
	content = append(content, byte('\n'))
	_, err = w.Write(content)
	if err != nil {
		return err
	}
	return w.Flush()
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
func ReceiveHelloMessage(r *bufio.Reader) (HelloMessage, error) {
	return ReceiveStandardMessage[HelloMessage](r, HelloMessageName)
}

// Transmit a HelloMessage over the channel, including name
func (msg HelloMessage) Transmit(w *bufio.Writer) error {
	return TransmitStandardMessage(w, HelloMessageName, msg)
}
