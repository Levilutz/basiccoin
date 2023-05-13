package main

import (
	"bufio"
	"fmt"

	"github.com/levilutz/basiccoin/src/util"
)

// HelloMessage

var HelloMessageName = "hello"

type HelloMessage struct {
	RuntimeID string `json:"runtimeID"`
	Version   string `json:"version"`
	Addr      string `json:"addr"`
}

// Receive the HelloMessage from the channel, consuming name if not done yet
func ReceiveHelloMessage(r *bufio.Reader, nameConsumed bool) (HelloMessage, error) {
	if !nameConsumed {
		data, err := util.RetryReadLine(r, 5)
		if err != nil {
			return HelloMessage{}, err
		}
		if string(data) != HelloMessageName {
			return HelloMessage{}, fmt.Errorf(
				"expected '%s', received '%s'", HelloMessageName, string(data),
			)
		}
	}
	data, err := util.RetryReadLine(r, 5)
	if err != nil {
		return HelloMessage{}, err
	}
	content, err := util.UnJsonB64[HelloMessage](data)
	if err != nil {
		return HelloMessage{}, err
	}
	return content, nil
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
