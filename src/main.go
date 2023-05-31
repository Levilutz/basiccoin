package main

import (
	"fmt"
	"time"

	"github.com/levilutz/basiccoin/src/db"
	"github.com/levilutz/basiccoin/src/manager"
	"github.com/levilutz/basiccoin/src/peer"
	"github.com/levilutz/basiccoin/src/util"
)

func printComputedConstants() {
	fmt.Println("CoinbaseVSize:", db.CoinbaseVSize())
	fmt.Println("MinNonCoinbaseVSize:", db.MinNonCoinbaseVSize())
	fmt.Println("BlockMaxTxs", db.BlockMaxTxs())
	fmt.Println("MerkleTreeMaxSize", db.MerkleTreeMaxSize())
}

func main() {
	cli_args := util.ParseCLIArgs()
	util.PrettyPrint(cli_args)
	util.PrettyPrint(util.Constants)
	printComputedConstants()

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

		// Set local addr if not set from args
		if cli_args.LocalAddr == "" {
			util.Constants.LocalAddr = pc.LocalAddr().IP.String() + ":21720"
			fmt.Println("Discovered address of", util.Constants.LocalAddr)
		}

		go manager.IntroducePeerConn(pc, true)
	}

	if cli_args.Listen {
		go manager.Listen()
	}
	manager.Loop()
}
