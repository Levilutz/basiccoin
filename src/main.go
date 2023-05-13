package main

import (
	"net"

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

	// Greet seed peer
	if cli_args.SeedAddr != "" {
		conn, err := peer.ResolvePeerConn(cli_args.SeedAddr)
		util.PanicErr(err)
		err = peer.GreetPeer(conn)
		util.PanicErr(err)
	}

	for {
		select {
		case conn := <-conns:
			// TODO make bus
			go peer.ReceivePeerGreeting(peer.NewPeerConn(conn))
		}
	}

	// TODO
	// Make main loop receive events from 1. listener channel and
	// 2. peer bus kill signals
	// 3. all peer buses (flattened?? and randomly shuffled per round)
	// On new conn, make bus and receive greeting
}
