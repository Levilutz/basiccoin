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
		data, err := util.RetryReadBytes(reader, 5)
		fmt.Println(data, err)
		done <- true
	}()

	time.Sleep(time.Second)

	addr, _ := net.ResolveTCPAddr("tcp", "0.0.0.0:21720")
	conn, _ := net.DialTCP("tcp", nil, addr)
	conn.Close()

	<-done
}
