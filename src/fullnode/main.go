package main

import (
	"fmt"
	"time"

	"github.com/levilutz/basiccoin/src/kern"
	"github.com/levilutz/basiccoin/src/peer"
	"github.com/levilutz/basiccoin/src/util"
)

func main() {
	cli_args := ParseCLIArgs()
	if util.Constants.DebugLevel >= 1 {
		fmt.Println(util.PrettyPrint(cli_args))
		fmt.Println(util.PrettyPrint(util.Constants))
	}
	fmt.Printf("starting with id %s\n", util.Constants.RuntimeID)

	manager := NewManager(cli_args.MinerPayoutAddr, kern.DevNetParams())

	if cli_args.SeedAddr != "" {
		var pc *peer.PeerConn
		var err error
		for i := 0; i < 5; i++ {
			pc, err = peer.ResolvePeerConn(cli_args.SeedAddr)
			if err == nil || i == 4 {
				break
			}
			fmt.Println("failed attempt to connect to seed peer")
			time.Sleep(5 * time.Second)
		}
		util.PanicErr(err)
		fmt.Println("successfully connected to seed peer")
		// Set local addr if not set from args
		if cli_args.LocalAddr == "" {
			util.Constants.LocalAddr = pc.LocalAddr().IP.String() + ":21720"
			fmt.Println("discovered own address of", util.Constants.LocalAddr)
		}

		go manager.IntroducePeerConn(pc, true)
	}

	if cli_args.Listen {
		go manager.Listen()
	}
	manager.Loop()
}
