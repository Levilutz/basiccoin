package main

import (
	"net"
	"sync"

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
	peerBusesMutex := sync.Mutex{}

	// Greet seed peer
	if cli_args.SeedAddr != "" {
		conn, err := peer.ResolvePeerConn(cli_args.SeedAddr)
		util.PanicErr(err)
		bus, err := peer.GreetPeer(conn, mainBus)
		util.PanicErr(err)
		peerBuses[bus.PeerRuntimeID] = bus
	}

	managerRoutine(MainState{
		newConnChannel: conns,
		mainBus:        mainBus,
		peerBuses:      peerBuses,
		peerBusesMutex: &peerBusesMutex,
	})
}
