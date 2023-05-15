package main

import (
	"net"

	"github.com/levilutz/basiccoin/src/events"
	"github.com/levilutz/basiccoin/src/manager"
	"github.com/levilutz/basiccoin/src/peer"
	"github.com/levilutz/basiccoin/src/util"
)

func main() {
	cli_args := util.ParseCLIArgs()
	util.PrettyPrint(cli_args)
	util.PrettyPrint(util.Constants)

	// Start listening for new peers
	var conns chan *net.TCPConn
	if util.Constants.Listen {
		conns = make(chan *net.TCPConn)
		go util.ListenTCP(conns)
	} else {
		conns = nil
	}

	// Buses
	mainBus := make(chan events.MainEvent)
	peers := make(map[string]*peer.Peer)
	knownPeerAddrs := make(map[string]struct{}, 0)

	// Greet seed peer
	if cli_args.SeedAddr != "" {
		p, err := peer.NewPeerOutbound(cli_args.SeedAddr, mainBus)
		util.PanicErr(err)
		go p.Loop()
		peers[p.HelloMsg.RuntimeID] = p
		knownPeerAddrs[cli_args.SeedAddr] = struct{}{}
	}

	manager := manager.NewManager(conns, mainBus, peers, knownPeerAddrs)
	manager.Loop()
}
