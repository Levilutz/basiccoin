package peer

import (
	"bufio"
	"encoding/binary"
	"encoding/hex"
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
	c *net.TCPConn
	r *bufio.Reader
	w *bufio.Writer
	e error
}

// Create a peer connection from a TCP Connection.
func NewPeerConn(c *net.TCPConn) *PeerConn {
	return &PeerConn{
		c: c,
		r: bufio.NewReader(c),
		w: bufio.NewWriter(c),
		e: nil,
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

// Get local address as seen by peer.
func (pc *PeerConn) LocalAddr() *net.TCPAddr {
	return pc.c.LocalAddr().(*net.TCPAddr)
}

func (pc *PeerConn) Handshake() *HelloMessage {
	if pc.e != nil {
		return nil
	}
	pc.TransmitStringLine("basiccoin")
	pc.ConsumeExpected("basiccoin")
	pc.TransmitMessage(NewHelloMessage())
	helloMsg, err := ReceiveHelloMessage(pc)
	if err != nil {
		pc.e = err
		return nil
	}
	return &helloMsg
}

func (pc *PeerConn) HandshakeHeights(height uint64) uint64 {
	if pc.e != nil {
		return 0
	}
	pc.TransmitUint64Line(height)
	theirHeight := pc.RetryReadUint64Line(7)
	return theirHeight
}

// Transmit continue|close, and receive their continue|close. Return nil if both peers
// want to connect, or a reason not to otherwise.
func (pc *PeerConn) VerifyConnWanted(msg HelloMessage) {
	if pc.e != nil {
		return
	}
	// Close if we don't want connection
	if msg.RuntimeID == util.Constants.RuntimeID ||
		msg.Version != util.Constants.Version {
		pc.TransmitStringLine("cmd:close")
		if pc.e != nil {
			return
		}
		if err := pc.c.Close(); err != nil {
			pc.e = err
			return
		}
		pc.e = errors.New("we do not want connection")
		return
	}

	pc.TransmitStringLine("cmd:continue")
	// Receive whether they want to continue
	contMsg := pc.RetryReadLine(7)
	if pc.e != nil {
		return
	} else if string(contMsg) == "cmd:continue" {
		return
	} else if string(contMsg) == "cmd:close" {
		if err := pc.c.Close(); err != nil {
			pc.e = err
			return
		}
		pc.e = errors.New("peer does not want connection")
	} else {
		pc.e = fmt.Errorf("expected continue|close, received '%s'", contMsg)
	}
}

// Consume the next line and assert that it matches msg.
// Do not include \n in msg.
func (pc *PeerConn) ConsumeExpected(msg string) {
	if pc.e != nil {
		return
	}
	data := pc.RetryReadLine(7)
	if pc.e != nil {
		return
	}
	if string(data) != msg {
		pc.e = fmt.Errorf(
			"expected '%s', received '%s'", msg, string(data),
		)
	}
}

// Transmit a Message.
func (pc *PeerConn) TransmitMessage(msg PeerMessage) {
	if pc.e != nil {
		return
	}
	pc.e = msg.Transmit(pc)
}

// Transmit bytes as a line.
// Do not include \n in msg.
func (pc *PeerConn) TransmitLine(msg []byte) {
	if pc.e != nil {
		return
	}
	if util.Constants.DebugNetwork {
		fmt.Println("NET_OUT", string(msg))
	}
	_, err := pc.w.Write(append(msg, byte('\n')))
	if err != nil {
		pc.e = err
		return
	}
	pc.e = pc.w.Flush()
}

// Transmit a string as a line.
// Do not include \n in msg.
func (pc *PeerConn) TransmitStringLine(msg string) {
	if pc.e != nil {
		return
	}
	pc.TransmitLine([]byte(msg))
}

// Transmit an int as a line.
func (pc *PeerConn) TransmitIntLine(msg int) {
	if pc.e != nil {
		return
	}
	pc.TransmitLine([]byte(strconv.Itoa(msg)))
}

// Transmit a uint64 as a line.
func (pc *PeerConn) TransmitUint64Line(msg uint64) {
	if pc.e != nil {
		return
	}
	bs := make([]byte, 8)
	binary.BigEndian.PutUint64(bs, msg)
	pc.TransmitStringLine(fmt.Sprintf("%x", bs))
}

// Retry reading a string line, exponential wait.
// See RetryReadLine for more info.
func (pc *PeerConn) RetryReadStringLine(attempts int) string {
	if pc.e != nil {
		return ""
	}
	raw := pc.RetryReadLine(attempts)
	if pc.e != nil {
		return ""
	}
	return string(raw)
}

// Retry reading an int line, exponential wait.
// See RetryReadLine for more info.
func (pc *PeerConn) RetryReadIntLine(attempts int) int {
	if pc.e != nil {
		return 0
	}
	raw := pc.RetryReadLine(attempts)
	if pc.e != nil {
		return 0
	}
	num, err := strconv.Atoi(string(raw))
	if err != nil {
		pc.e = err
		return 0
	}
	return num
}

// Retry reading a uint64 line, exponential wait.
// See RetryReadLine for more info.
func (pc *PeerConn) RetryReadUint64Line(attempts int) uint64 {
	if pc.e != nil {
		return 0
	}
	raw := pc.RetryReadLine(attempts)
	if pc.e != nil {
		return 0
	}
	out, err := hex.DecodeString(string(raw))
	if err != nil {
		pc.e = err
		return 0
	}
	return binary.BigEndian.Uint64(out)
}

// Retry reading a line, exponential wait.
// Attempt delays begin at 100ms and multiply by 2.
// Estimated max total runtime = (2^attempts - 1) * 0.1 seconds.
func (pc *PeerConn) RetryReadLine(attempts int) []byte {
	if pc.e != nil {
		return nil
	}
	delay := time.Duration(100) * time.Millisecond
	for i := 0; i < attempts; i++ {
		data := pc.ReadLineTimeout(delay)
		if pc.e == nil {
			return data
		} else if errors.Is(pc.e, io.EOF) || errors.Is(pc.e, os.ErrDeadlineExceeded) {
			pc.e = nil
			delay *= time.Duration(2)
		} else {
			return nil
		}
	}
	pc.e = io.EOF
	return nil
}

// Attempt to read a line, with timeout
func (pc *PeerConn) ReadLineTimeout(timeout time.Duration) []byte {
	if pc.e != nil {
		return nil
	}
	defer pc.c.SetReadDeadline(time.Time{})
	pc.c.SetReadDeadline(time.Now().Add(timeout))
	data, err := pc.r.ReadBytes(byte('\n'))
	if err != nil {
		pc.e = err
		return nil
	}
	data = data[:len(data)-1] // len(data) will always be at least 1
	if util.Constants.DebugNetwork {
		fmt.Println("NET_IN", string(data))
	}
	return data
}

// Close the connection.
func (pc *PeerConn) Close() error {
	return pc.c.Close()
}

// Check whether we have a stored error.
func (pc *PeerConn) HasErr() bool {
	return pc.e != nil
}

// Pop the stored error
func (pc *PeerConn) Err() error {
	defer func() { pc.e = nil }()
	return pc.e
}
