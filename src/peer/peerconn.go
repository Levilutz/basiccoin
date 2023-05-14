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

// Encapsulate a low-level connection to peer.
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

// Give an initial connection handshake
func (pc *PeerConn) GiveHandshake() *HelloPeerMessage {
	pc.TransmitMessage(NewHelloMessage())
	pc.ConsumeExpected("ack:hello")
	pc.ConsumeExpected("hello")
	if pc.E != nil {
		return nil
	}
	helloMsg, err := ReceiveHelloMessage(pc)
	if err != nil {
		pc.E = err
		return nil
	}
	pc.TransmitStringLine("ack:hello")
	if pc.E != nil {
		return nil
	}
	return &helloMsg
}

// Receive an initial connection handshake
func (pc *PeerConn) ReceiveHandshake() *HelloPeerMessage {
	pc.ConsumeExpected("hello")
	if pc.E != nil {
		return nil
	}
	helloMsg, err := ReceiveHelloMessage(pc)
	if err != nil {
		pc.E = err
		return nil
	}
	pc.TransmitStringLine("ack:hello")
	pc.TransmitMessage(NewHelloMessage())
	pc.ConsumeExpected("ack:hello")
	if pc.E != nil {
		return nil
	}
	return &helloMsg
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
	delay := time.Duration(100) * time.Millisecond
	for i := 0; i < attempts; i++ {
		data := pc.ReadLineTimeout(delay)
		if pc.E == nil {
			return data[:len(data)-1] // len(data) will always be at least 1
		} else if errors.Is(pc.E, io.EOF) || errors.Is(pc.E, os.ErrDeadlineExceeded) {
			pc.E = nil
			delay *= time.Duration(2)
		} else {
			return nil
		}
	}
	pc.E = io.EOF
	return nil
}

// Attempt to read a line, with timeout
func (pc *PeerConn) ReadLineTimeout(timeout time.Duration) []byte {
	if pc.E != nil {
		return nil
	}
	defer pc.C.SetReadDeadline(time.Time{})
	pc.C.SetReadDeadline(time.Now().Add(timeout))
	data, err := pc.R.ReadBytes(byte('\n'))
	if err != nil {
		pc.E = err
		return nil
	}
	return data
}

// Pop the stored error
func (pc *PeerConn) Err() error {
	defer func() { pc.E = nil }()
	return pc.E
}
