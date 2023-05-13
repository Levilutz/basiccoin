package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/levilutz/basiccoin/src/util"
)

func main() {
	cli_args := GetCLIArgs()
	util.PrettyPrint(cli_args)
	util.PrettyPrint(util.Constants)

	if cli_args.SeedAddr == "" {
		addr, _ := net.ResolveTCPAddr("tcp", "0.0.0.0:21720")
		listen, _ := net.ListenTCP("tcp", addr)
		defer listen.Close()
		conn, _ := listen.Accept()
		defer conn.Close()
		reader := bufio.NewReader(conn)
		for {
			data, err := reader.ReadBytes(byte('\n'))
			if errors.Is(err, io.EOF) {
				continue
			} else if err != nil {
				fmt.Println(err)
			} else if string(data) == "close\n" {
				fmt.Println("happy close")
				return
			} else {
				fmt.Print(string(data))
			}
		}
	} else {
		a, _ := net.ResolveTCPAddr("tcp", cli_args.SeedAddr)
		c, _ := net.DialTCP("tcp", nil, a)
		defer c.Close()
		for i := 0; i < 10; i++ {
			c.Write([]byte("bogus\n"))
		}
		c.Write([]byte("close\n"))
	}
}
