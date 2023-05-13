package main

import (
	"net"

	"github.com/levilutz/basiccoin/src/util"
)

func main() {
	cli_args := util.ParseCLIArgs()
	util.PrettyPrint(cli_args)
	util.PrettyPrint(util.Constants)

	_, err := net.ResolveTCPAddr("tcp", cli_args.LocalAddr)
	if err != nil {
		panic(err)
	}
}
