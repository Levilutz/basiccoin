package prot

import (
	"encoding/binary"
	"fmt"
)

// Read a string from the conn.
func (c *Conn) ReadString() string {
	if c.err != nil {
		return ""
	}
	raw := c.Read()
	if c.err != nil {
		return ""
	}
	return string(raw)
}

// Read an expected string from the conn, err if read string was different.
func (c *Conn) ReadStringExpected(expected string) {
	if c.err != nil {
		return
	}
	actual := c.ReadString()
	if c.err != nil {
		return
	}
	if actual != expected {
		c.err = fmt.Errorf("received incorrect string: %s != %s", actual, expected)
	}
}

// Write a string to the conn.
func (c *Conn) WriteString(data string) {
	if c.err != nil {
		return
	}
	c.Write([]byte(data))
}

// Read a Uint64 from the conn.
func (c *Conn) ReadUint64() uint64 {
	if c.err != nil {
		return 0
	}
	raw := c.readRawTimeout(8, defaultTimeout)
	if c.err != nil {
		return 0
	}
	return binary.BigEndian.Uint64(raw)
}

// Write a Uint64 to the conn.
func (c *Conn) WriteUint64(data uint64) {
	if c.err != nil {
		return
	}
	dataB := make([]byte, 8)
	binary.BigEndian.PutUint64(dataB, data)
	c.writeRawTimeout(dataB, defaultTimeout)
}

// Read a bool from the conn.
func (c *Conn) ReadBool() bool {
	if c.err != nil {
		return false
	}
	raw := c.readRawTimeout(1, defaultTimeout)
	if c.err != nil {
		return false
	}
	if raw[0] == byte(0) {
		return false
	} else if raw[0] == byte(1) {
		return true
	} else {
		c.err = fmt.Errorf("unrecognized bool byte: %d", raw[0])
		return false
	}
}

// Write a bool to the conn.
func (c *Conn) WriteBool(data bool) {
	if c.err != nil {
		return
	}
	if data {
		c.writeRawTimeout([]byte{1}, defaultTimeout)
	} else {
		c.writeRawTimeout([]byte{0}, defaultTimeout)
	}
}
