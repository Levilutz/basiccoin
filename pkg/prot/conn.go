package prot

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

const defaultTimeout = time.Second * 30

// A low-level connection to a peer.
type Conn struct {
	tc  *net.TCPConn
	err error
}

// Create a new connection from an existing TCP Connection.
// The returned conn might have an err set, but this won't return one directly.
func NewConn(tcpConn *net.TCPConn) *Conn {
	conn := &Conn{
		tc:  tcpConn,
		err: nil,
	}
	conn.handshake()
	return conn
}

// Initial handshake for an incompletely-initialized conn.
func (c *Conn) handshake() {
	if c.err != nil {
		return
	}
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

// Read data from the conn with the given timeout for the size bytes, then the default.
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

// Read data from the conn with the default timeout for each read.
func (c *Conn) Read() []byte {
	if c.err != nil {
		return nil
	}
	return c.ReadTimeout(defaultTimeout)
}

// Write data to the conn with the default timeout for each write.
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

// Check whether we have a stored error.
func (c *Conn) HasErr() bool {
	return c.err != nil
}

// Pop the stored error.
func (c *Conn) Err() error {
	defer func() { c.err = nil }()
	return c.err
}
