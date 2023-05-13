package main

import (
	"bufio"
	"fmt"
	"net"
	"time"

	"github.com/levilutz/basiccoin/src/util"
)

func main() {
	cli_args := GetCLIArgs()
	util.PrettyPrint(cli_args)
	util.PrettyPrint(util.Constants)

	done := make(chan bool)

	go func() {
		addr, _ := net.ResolveTCPAddr("tcp", "0.0.0.0:21720")
		listen, _ := net.ListenTCP("tcp", addr)
		defer listen.Close()
		conn, _ := listen.Accept()
		defer conn.Close()
		reader := bufio.NewReader(conn)
		msg, err := ReceiveHelloMessage(reader, false)
		if err != nil {
			fmt.Println(err)
		} else {
			util.PrettyPrint(msg)
		}
		done <- true
	}()

	time.Sleep(time.Second)

	addr, _ := net.ResolveTCPAddr("tcp", "0.0.0.0:21720")
	conn, _ := net.DialTCP("tcp", nil, addr)
	writer := bufio.NewWriter(conn)
	msg := HelloMessage{
		RuntimeID: util.Constants.RuntimeID,
		Version:   util.Constants.Version,
		Addr:      "",
	}
	msg.Transmit(writer)
	conn.Close()

	<-done
}
