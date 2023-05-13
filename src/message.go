package main

import (
	"bufio"
	"fmt"

	"github.com/levilutz/basiccoin/src/util"
)

// Generic helpers

// Consume the next line and assert that it matches msg
func ConsumeExpected(r *bufio.Reader, msg string) error {
	data, err := util.RetryReadLine(r, 5)
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

// Receive base64(json(message)) from a single line
func ReceiveStandardMessage[R any](msgName string, r *bufio.Reader) (R, error) {
	var content R
	data, err := util.RetryReadLine(r, 5)
	if err != nil {
		return content, err
	}
	return util.UnJsonB64[R](data)
}

// Transmit msgName then base64(json(message)) in a single line each
func TransmitStandardMessage(msgName string, msg any, w *bufio.Writer) error {
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

// Receive the HelloMessage from the channel
func ReceiveHelloMessage(r *bufio.Reader) (HelloMessage, error) {
	return ReceiveStandardMessage[HelloMessage](HelloMessageName, r)
}

// Transmit the HelloMessage over the channel, including name
func (msg HelloMessage) Transmit(w *bufio.Writer) error {
	return TransmitStandardMessage(HelloMessageName, msg, w)
}
