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
	E error
}

// Create a peer connection from a TCP Connection.
func NewPeerConn(c *net.TCPConn) *PeerConn {
	return &PeerConn{
		C: c,
		R: bufio.NewReader(c),
		W: bufio.NewWriter(c),
		E: nil,
	}
}

// Resolve and dial a TCP Address then make a peer connection if successful.
func ResolvePeerConn(addr string) (*PeerConn, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, err
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return nil, err
	}
	return NewPeerConn(conn), nil
}

// Consume the next line and assert that it matches msg.
// Do not include \n in msg.
func (pc *PeerConn) ConsumeExpected(msg string) {
	if pc.E != nil {
		return
	}
	data := pc.RetryReadLine(7)
	if pc.E != nil {
		return
	}
	if string(data) != msg {
		pc.E = fmt.Errorf(
			"expected '%s', received '%s'", msg, string(data),
		)
	}
}

// Transmit a Message.
func (pc *PeerConn) TransmitMessage(msg Message) {
	if pc.E != nil {
		return
	}
	pc.E = msg.Transmit(pc)
}

// Transmit a simple string as a line.
// Do not include \n in msg.
func (pc *PeerConn) TransmitStringLine(msg string) {
	if pc.E != nil {
		return
	}
	_, err := pc.W.Write([]byte(msg + "\n"))
	if err != nil {
		pc.E = err
		return
	}
	pc.E = pc.W.Flush()
}

// Receive base64(json(message)) from a single line
func PeerConnReceiveStandardMessage[R Message](pc *PeerConn) R {
	// Cannot be method until golang allows type params on methods
	var content R
	if pc.E != nil {
		return content
	}
	data := pc.RetryReadLine(7)
	if pc.E != nil {
		return content
	}
	content, err := util.UnJsonB64[R](data)
	if err != nil {
		pc.E = err
	}
	return content
}

// Transmit msgName then base64(json(message)) in a single line each
func (pc *PeerConn) TransmitStandardMessage(msg Message) {
	if pc.E != nil {
		return
	}
	data, err := util.JsonB64(msg)
	if err != nil {
		pc.E = err
		return
	}
	content := []byte(msg.GetName() + "\n")
	content = append(content, data...)
	content = append(content, byte('\n'))
	_, err = pc.W.Write(content)
	if err != nil {
		pc.E = err
		return
	}
	pc.E = pc.W.Flush()
}

// Retry reading a line, exponential wait.
// Attempt delays begin at 100ms and multiply by 2.
// Max total runtime: 1 > 100ms, 2 > 300ms, 3 > 700ms, 4 > 1.5s, 5 > 3.1s, 6 > 6.3s,
// 7 > 12.7s, 8 > 25.5s, 9 > 51.1s, 10 > 102.3s, etc.
func (pc *PeerConn) RetryReadLine(attempts int) []byte {
	if pc.E != nil {
		return nil
	}
	defer pc.C.SetReadDeadline(time.Time{})
	delay := time.Duration(100) * time.Millisecond
	for i := 0; i < attempts; i++ {
		pc.C.SetReadDeadline(time.Now().Add(delay))
		data, err := pc.R.ReadBytes(byte('\n'))
		if err == nil {
			if len(data) > 0 {
				return data[:len(data)-1]
			} else {
				return data
			}
		} else if (errors.Is(err, io.EOF) || errors.Is(err, os.ErrDeadlineExceeded)) &&
			i != attempts-1 {
			delay *= time.Duration(2)
			continue
		} else {
			pc.E = err
			return nil
		}
	}
	pc.E = io.EOF
	return nil
}

// Pop the stored error
func (pc *PeerConn) Err() error {
	defer func() { pc.E = nil }()
	return pc.E
}
