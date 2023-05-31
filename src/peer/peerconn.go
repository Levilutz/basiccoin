package peer

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/levilutz/basiccoin/src/util"
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

// Verify that a peer exists on this connection, then close.
func (pc *PeerConn) VerifyAndClose() {
	if pc.E != nil {
		return
	}
	helloMsg := pc.Handshake()
	pc.TransmitStringLine("cmd:close")
	if pc.E != nil {
	} else if err := pc.C.Close(); err != nil {
		pc.E = err
	} else if helloMsg.RuntimeID == util.Constants.RuntimeID {
		pc.E = errors.New("peer is us")
	}
}

func (pc *PeerConn) Handshake() *HelloMessage {
	if pc.E != nil {
		return nil
	}
	pc.TransmitStringLine("basiccoin")
	pc.ConsumeExpected("basiccoin")
	pc.TransmitMessage(NewHelloMessage())
	helloMsg, err := ReceiveHelloMessage(pc)
	if err != nil {
		pc.E = err
		return nil
	}
	return &helloMsg
}

// Transmit continue|close, and receive their continue|close. Return nil if both peers
// want to connect, or a reason not to otherwise.
func (pc *PeerConn) VerifyConnWanted(msg HelloMessage) {
	if pc.E != nil {
		return
	}
	// Close if we don't want connection
	if msg.RuntimeID == util.Constants.RuntimeID ||
		msg.Version != util.Constants.Version {
		pc.TransmitStringLine("cmd:close")
		if pc.E != nil {
			return
		}
		if err := pc.C.Close(); err != nil {
			pc.E = err
			return
		}
		pc.E = errors.New("we do not want connection")
		return
	}

	pc.TransmitStringLine("cmd:continue")
	// Receive whether they want to continue
	contMsg := pc.RetryReadLine(7)
	if pc.E != nil {
		return
	} else if string(contMsg) == "cmd:continue" {
		return
	} else if string(contMsg) == "cmd:close" {
		if err := pc.C.Close(); err != nil {
			pc.E = err
			return
		}
		pc.E = errors.New("peer does not want connection")
	} else {
		pc.E = fmt.Errorf("expected continue|close, received '%s'", contMsg)
	}
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
	if util.Constants.DebugNetwork {
		fmt.Println("NET_OUT", string(msg))
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
	if pc.E != nil {
		return
	}
	pc.TransmitLine([]byte(msg))
}

// Transmit an int as a line.
func (pc *PeerConn) TransmitIntLine(msg int) {
	if pc.E != nil {
		return
	}
	pc.TransmitLine([]byte(strconv.Itoa(msg)))
}

// Retry reading a string line, exponential wait.
// See RetryReadLine for more info.
func (pc *PeerConn) RetryReadStringLine(attempts int) string {
	if pc.E != nil {
		return ""
	}
	raw := pc.RetryReadLine(attempts)
	if pc.E != nil {
		return ""
	}
	return string(raw)
}

// Retry reading an int line, exponential wait.
// See RetryReadLine for more info.
func (pc *PeerConn) RetryReadIntLine(attempts int) int {
	if pc.E != nil {
		return 0
	}
	raw := pc.RetryReadLine(attempts)
	if pc.E != nil {
		return 0
	}
	num, err := strconv.Atoi(string(raw))
	if err != nil {
		pc.E = err
		return 0
	}
	return num
}

// Retry reading a line, exponential wait.
// Attempt delays begin at 100ms and multiply by 2.
// Estimated max total runtime = (2^attempts - 1) * 0.1 seconds.
func (pc *PeerConn) RetryReadLine(attempts int) []byte {
	if pc.E != nil {
		return nil
	}
	delay := time.Duration(100) * time.Millisecond
	for i := 0; i < attempts; i++ {
		data := pc.ReadLineTimeout(delay)
		if pc.E == nil {
			return data
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
	data = data[:len(data)-1] // len(data) will always be at least 1
	if util.Constants.DebugNetwork {
		fmt.Println("NET_IN", string(data))
	}
	return data
}

// Pop the stored error
func (pc *PeerConn) Err() error {
	defer func() { pc.E = nil }()
	return pc.E
}
