package main

import (
	"fmt"
	"time"

	"github.com/levilutz/basiccoin/src/manager"
	"github.com/levilutz/basiccoin/src/peer"
	"github.com/levilutz/basiccoin/src/util"
)

func main() {
	cli_args := util.ParseCLIArgs()
	util.PrettyPrint(cli_args)
	util.PrettyPrint(util.Constants)

	manager := manager.NewManager()

	if cli_args.SeedAddr != "" {
		var pc *peer.PeerConn
		var err error
		for i := 0; i < 5; i++ {
			pc, err = peer.ResolvePeerConn(cli_args.SeedAddr)
			if err == nil || i == 4 {
				break
			}
			fmt.Println("Failed attempt to contact seed peer")
			time.Sleep(5 * time.Second)
		}
		util.PanicErr(err)
		go manager.IntroducePeerConn(pc, true)
	}

	if cli_args.Listen {
		go manager.Listen()
	}
	manager.Loop()
}
