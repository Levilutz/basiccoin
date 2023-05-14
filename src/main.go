package main

import (
	"net"

	"github.com/levilutz/basiccoin/src/mainbus"
	"github.com/levilutz/basiccoin/src/peer"
	"github.com/levilutz/basiccoin/src/util"
)

func main() {
	cli_args := util.ParseCLIArgs()
	util.PrettyPrint(cli_args)
	util.PrettyPrint(util.Constants)

	// Start listening for new peers
	conns := make(chan *net.TCPConn)
	go util.ListenTCP(conns)

	// Buses
	mainBus := mainbus.NewMainBus(100)
	peers := make(map[string]*peer.Peer)

	// Greet seed peer
	if cli_args.SeedAddr != "" {
		p, err := peer.NewPeerOutbound(cli_args.SeedAddr, mainBus)
		util.PanicErr(err)
		go p.Loop()
		peers[p.HelloMsg.RuntimeID] = p
	}

	managerRoutine(&MainState{
		newConnChannel:     conns,
		mainBus:            mainBus,
		peers:              peers,
		candidatePeerAddrs: make(map[string]struct{}, 0),
	})
}
