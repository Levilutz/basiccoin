package main

import (
	"bufio"
	"net"

	"github.com/levilutz/basiccoin/src/util"
)

func PeerRoutine(c *net.TCPConn) {
	r := bufio.NewReader(c)
	w := bufio.NewWriter(c)
	err := NewHelloMessage().Transmit(w)
	util.PanicErr(err)
	err = ConsumeExpected(r, "ack:hello")
	util.PanicErr(err)
}
