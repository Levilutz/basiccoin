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

	"github.com/levilutz/basiccoin/src/db"
	"github.com/levilutz/basiccoin/src/util"
)

// Basic peer data exchanged in a handshake.
type PeerInfo struct {
	RuntimeID string
	Version   string
	Addr      string
}

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

// Exchange basic information about each other
func (pc *PeerConn) Handshake() *PeerInfo {
	if pc.e != nil {
		return nil
	}
	pc.TransmitStringLine("basiccoin")
	pc.ConsumeExpected("basiccoin")
	// Transmit basic info
	pc.TransmitStringLine(util.Constants.RuntimeID)
	pc.TransmitStringLine(util.Constants.Version)
	if util.Constants.Listen {
		pc.TransmitStringLine(util.Constants.LocalAddr)
	} else {
		pc.TransmitStringLine("")
	}
	// Receive basic info
	info := PeerInfo{
		RuntimeID: pc.RetryReadStringLine(7),
		Version:   pc.RetryReadStringLine(7),
		Addr:      pc.RetryReadStringLine(7),
	}
	if pc.e != nil {
		return nil
	}
	return &info
}

// Transmit continue|close, and receive their continue|close. Return nil if both peers
// want to connect, or a reason not to otherwise.
func (pc *PeerConn) VerifyConnWanted(info PeerInfo) {
	if pc.e != nil {
		return
	}
	// Close if we don't want connection
	if info.RuntimeID == util.Constants.RuntimeID ||
		info.Version != util.Constants.Version {
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

// Transmit bytes as a line.
// Do not include \n in msg.
func (pc *PeerConn) TransmitLine(msg []byte) {
	if pc.e != nil {
		return
	}
	if util.Constants.DebugLevel >= 2 {
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

// Transmit a hash as a line.
func (pc *PeerConn) TransmitHashLine(msg db.HashT) {
	if pc.e != nil {
		return
	}
	pc.TransmitStringLine(fmt.Sprintf("%x", msg))
}

// Transmit bytes as a hex line.
func (pc *PeerConn) TransmitBytesHexLine(msg []byte) {
	if pc.e != nil {
		return
	}
	pc.TransmitStringLine(fmt.Sprintf("%x", msg))
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

// Retry reading a hash line, exponential wait.
// See RetryReadLine for more info.
func (pc *PeerConn) RetryReadHashLine(attempts int) db.HashT {
	if pc.e != nil {
		return db.HashTZero
	}
	raw := pc.RetryReadLine(attempts)
	if pc.e != nil {
		return db.HashTZero
	}
	out, err := hex.DecodeString(string(raw))
	if err != nil {
		pc.e = err
		return db.HashTZero
	}
	if len(out) != 32 {
		pc.e = fmt.Errorf("cannot decode hash - unexpected length %d", len(out))
		return db.HashTZero
	}
	return db.HashT(out)
}

// Retry reading a bytes line as hex, exponential wait.
// See RetryReadLine for more info.
func (pc *PeerConn) RetryReadBytesHexLine(attempts int) []byte {
	if pc.e != nil {
		return nil
	}
	raw := pc.RetryReadLine(attempts)
	if pc.e != nil {
		return nil
	}
	out, err := hex.DecodeString(string(raw))
	if err != nil {
		pc.e = err
		return nil
	}
	return out
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
	if util.Constants.DebugLevel >= 2 {
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
