package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"github.com/levilutz/basiccoin/src/util"
)

// Encapsulate a successful initial connection to peer
type PeerConn struct {
	C *net.TCPConn
	R *bufio.Reader
	W *bufio.Writer
}

// Create a peer connection from a TCP Connection.
func NewPeerConn(c *net.TCPConn) PeerConn {
	return PeerConn{
		C: c,
		R: bufio.NewReader(c),
		W: bufio.NewWriter(c),
	}
}

// Dial a TCP Address and make a peer connection if successful.
func ResolvePeerConn(addr string) (PeerConn, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return PeerConn{}, err
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return PeerConn{}, err
	}
	return NewPeerConn(conn), nil
}

// Consume the next line and assert that it matches msg.
// Do not include \n in msg.
func (pc PeerConn) ConsumeExpected(msg string) error {
	data, err := pc.RetryReadLine(7)
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

// Transmit a simple string as a line.
// Do not include \n in msg.
func (pc PeerConn) TransmitStringLine(msg string) error {
	_, err := pc.W.Write([]byte(msg + "\n"))
	if err != nil {
		return err
	}
	return pc.W.Flush()
}

// Receive base64(json(message)) from a single line
func PeerConnReceiveStandardMessage[R Message](pc PeerConn) (R, error) {
	// Cannot be method until golang allows type params on methods
	var content R
	data, err := pc.RetryReadLine(7)
	if err != nil {
		return content, err
	}
	return util.UnJsonB64[R](data)
}

// Transmit msgName then base64(json(message)) in a single line each
func (pc PeerConn) TransmitStandardMessage(msg Message) error {
	data, err := util.JsonB64(msg)
	if err != nil {
		return err
	}
	content := []byte(msg.GetName() + "\n")
	content = append(content, data...)
	content = append(content, byte('\n'))
	_, err = pc.W.Write(content)
	if err != nil {
		return err
	}
	return pc.W.Flush()
}

// Retry reading a line, exponential wait.
// Attempt delays begin at 100ms and multiply by 2.
// Max total runtime: 1 > 100ms, 2 > 300ms, 3 > 700ms, 4 > 1.5s, 5 > 3.1s, 6 > 6.3s,
// 7 > 12.7s, 8 > 25.5s, 9 > 51.1s, 10 > 102.3s, etc.
func (pc PeerConn) RetryReadLine(attempts int) ([]byte, error) {
	defer pc.C.SetReadDeadline(time.Time{})
	delay := time.Duration(100) * time.Millisecond
	for i := 0; i < attempts; i++ {
		pc.C.SetReadDeadline(time.Now().Add(delay))
		data, err := pc.R.ReadBytes(byte('\n'))
		if err == nil {
			if len(data) > 0 {
				return data[:len(data)-1], nil
			} else {
				return data, nil
			}
		} else if (errors.Is(err, io.EOF) || errors.Is(err, os.ErrDeadlineExceeded)) &&
			i != attempts-1 {
			delay *= time.Duration(2)
			continue
		} else {
			return nil, err
		}
	}
	return nil, io.EOF
}
