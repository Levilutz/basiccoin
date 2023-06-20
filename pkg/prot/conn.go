package prot

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"time"
)

const defaultTimeout = time.Second * 30

// A low-level connection to a peer.
type Conn struct {
	params        Params
	tc            *net.TCPConn
	peerRuntimeId string
	err           error
}

// Create a new connection from an existing TCP Connection.
// The returned conn might have an err set, but this won't return one directly.
func NewConn(params Params, tcpConn *net.TCPConn) *Conn {
	conn := &Conn{
		params: params,
		tc:     tcpConn,
		err:    nil,
		// peerRuntimeId is initialized by handshake
	}
	conn.handshake()
	return conn
}

func ResolveConn(params Params, addr string) (*Conn, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, err
	}
	tcpConn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return nil, err
	}
	return NewConn(params, tcpConn), nil
}

// Initial handshake for an incompletely-initialized conn.
func (c *Conn) handshake() {
	if c.err != nil {
		return
	}
	// Transmit handshake
	c.WriteString("levilutz/basiccoin")
	c.WriteString("v0.0.0") // If protocol is ever versioned this'll matter
	c.WriteString(c.params.RuntimeID)
	// Receive handshake
	c.ReadStringExpected("levilutz/basiccoin")
	c.ReadStringExpected("v0.0.0")
	peerRuntimeId := c.ReadString()
	// Cancel or continue the connection
	if c.err != nil {
		return
	}
	if peerRuntimeId == c.params.RuntimeID {
		c.WriteString("cancel")
		if c.err == nil {
			c.err = fmt.Errorf("will not connect to self")
		}
		c.Close()
		return
	}
	c.WriteString("continue")
	c.peerRuntimeId = peerRuntimeId
	// Handle the peer's desire to cancel or continue
	peerWants := c.ReadString()
	if c.err != nil {
		if os.IsTimeout(c.err) {
			c.Close()
		}
	} else if peerWants == "continue" {
		return
	} else if peerWants == "cancel" {
		c.err = fmt.Errorf("peer does not want connection")
		c.Close()
	} else {
		c.err = fmt.Errorf("unrecognized response: %s", peerWants)
		c.Close()
	}
}

// Get the runtime id of the peer.
func (c *Conn) PeerRuntimeId() string {
	return c.peerRuntimeId
}

// Get whether we initiated the connection.
func (c *Conn) WeAreInitiator() bool {
	return c.params.WeAreInitiator
}

// Get our local address as seen by the peer.
func (c *Conn) LocalAddr() *net.TCPAddr {
	return c.tc.LocalAddr().(*net.TCPAddr)
}

// Read the given number of bytes from the conn with the given timeout.
func (c *Conn) readRawTimeout(numBytes uint16, timeout time.Duration) []byte {
	if c.err != nil {
		return nil
	}
	c.tc.SetReadDeadline(time.Now().Add(timeout))
	defer c.tc.SetReadDeadline(time.Time{})
	data := make([]byte, numBytes)
	_, err := c.tc.Read(data)
	if err != nil {
		c.err = err
		return nil
	}
	return data
}

// Write the given bytes to the conn with the given timeout.
// The receiver should know exactly many bytes will be sent.
func (c *Conn) writeRawTimeout(data []byte, timeout time.Duration) {
	if c.err != nil {
		return
	}
	if len(data) > 65536 {
		c.err = fmt.Errorf("too many bytes to write: %d > 65536", len(data))
		return
	}
	c.tc.SetReadDeadline(time.Now().Add(timeout))
	defer c.tc.SetReadDeadline(time.Time{})
	_, err := c.tc.Write(data)
	if err != nil {
		c.err = err
	}
}

// Read variable-length data from the conn.
// Uses the given timeout for the size bytes, then the default for the data bytes.
func (c *Conn) ReadTimeout(timeout time.Duration) []byte {
	if c.err != nil {
		return nil
	}
	sizeB := c.readRawTimeout(2, timeout)
	if c.err != nil {
		return nil
	}
	size := binary.BigEndian.Uint16(sizeB)
	return c.readRawTimeout(size, defaultTimeout)
}

// Read variable-length data from the conn with the default timeout for each read.
func (c *Conn) Read() []byte {
	if c.err != nil {
		return nil
	}
	return c.ReadTimeout(defaultTimeout)
}

// Write variable-length data to the conn with the default timeout for each write.
func (c *Conn) Write(data []byte) {
	if c.err != nil {
		return
	}
	if len(data) > 65536 {
		c.err = fmt.Errorf("too many bytes to write: %d > 65536", len(data))
		return
	}
	sizeB := make([]byte, 2)
	binary.BigEndian.PutUint16(sizeB, uint16(len(data)))
	c.writeRawTimeout(sizeB, defaultTimeout)
	c.writeRawTimeout(data, defaultTimeout)
}

// Close the connection.
func (c *Conn) Close() error {
	return c.tc.Close()
}

// Try to close the connection, but don't care if it fails.
func (c *Conn) CloseIfPossible() {
	go func() {
		defer func() { recover() }()
		c.WriteString("cmd:close")
		c.Close()
	}()
}

// Check whether we have a stored error.
func (c *Conn) HasErr() bool {
	return c.err != nil
}

// Pop the stored error.
func (c *Conn) Err() error {
	defer func() { c.err = nil }()
	return c.err
}

// Pop the stored error, and panic if it wasn't a timeout.
func (c *Conn) TimeoutErrOrPanic() error {
	err := c.Err()
	if err != nil && !os.IsTimeout(err) {
		panic(err)
	}
	return err
}
