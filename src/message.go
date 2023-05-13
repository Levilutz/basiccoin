package main

import (
	"bufio"
	"fmt"

	"github.com/levilutz/basiccoin/src/util"
)

// Generic helpers

// Receive messageName then base64(json(message)) in a single line
func ReceiveSingleStandardMessage[R any](
	messageName string,
	r *bufio.Reader,
	nameConsumed bool,
) (R, error) {
	var content R
	if !nameConsumed {
		data, err := util.RetryReadLine(r, 5)
		if err != nil {
			return content, err
		}
		if string(data) != HelloMessageName {
			return content, fmt.Errorf(
				"expected '%s', received '%s'", messageName, string(data),
			)
		}
	}
	data, err := util.RetryReadLine(r, 5)
	if err != nil {
		return content, err
	}
	return util.UnJsonB64[R](data)
}

// HelloMessage

var HelloMessageName = "hello"

type HelloMessage struct {
	RuntimeID string `json:"runtimeID"`
	Version   string `json:"version"`
	Addr      string `json:"addr"`
}

// Receive the HelloMessage from the channel, consuming name if not done yet
func ReceiveHelloMessage(r *bufio.Reader, nameConsumed bool) (HelloMessage, error) {
	return ReceiveSingleStandardMessage[HelloMessage](
		HelloMessageName, r, nameConsumed,
	)
}

// Transmit the HelloMessage over the channel, including name
func (msg HelloMessage) Transmit(w *bufio.Writer) error {
	data, err := util.JsonB64(msg)
	if err != nil {
		return err
	}
	content := []byte(HelloMessageName + "\n")
	content = append(content, data...)
	content = append(content, byte('\n'))
	_, err = w.Write(content)
	if err != nil {
		return err
	}
	return w.Flush()
}
