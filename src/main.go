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
	peerBuses := make(map[string]*peer.PeerBus)

	// Greet seed peer
	if cli_args.SeedAddr != "" {
		conn, err := peer.ResolvePeerConn(cli_args.SeedAddr)
		util.PanicErr(err)
		msg, bus, err := peer.GreetPeer(conn, mainBus)
		util.PanicErr(err)
		peerBuses[msg.RuntimeID] = bus
	}

	managerRoutine(MainState{
		newConnChannel:     conns,
		mainBus:            mainBus,
		peerBuses:          peerBuses,
		candidatePeerAddrs: make(map[string]struct{}, 0),
	})
}
