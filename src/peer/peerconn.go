package peer

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"time"
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
func (pc *PeerConn) TransmitMessage(msg PeerMessage) {
	if pc.E != nil {
		return
	}
	pc.E = msg.Transmit(pc)
}

// Transmit bytes as a line.
// Do not include \n in msg.
func (pc *PeerConn) TransmitLine(msg []byte) {
	if pc.E != nil {
		return
	}
	_, err := pc.W.Write(append(msg, byte('\n')))
	if err != nil {
		pc.E = err
		return
	}
	pc.E = pc.W.Flush()
}

// Transmit a string as a line.
// Do not include \n in msg.
func (pc *PeerConn) TransmitStringLine(msg string) {
	pc.TransmitLine([]byte(msg))
}

// Retry reading a line, exponential wait.
// Attempt delays begin at 100ms and multiply by 2.
// Estimated max total runtime = (2^attempts - 1) * 0.1 seconds
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
			return data[:len(data)-1] // len(data) will always be at least 1
		} else if errors.Is(err, io.EOF) || errors.Is(err, os.ErrDeadlineExceeded) {
			delay *= time.Duration(2)
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
