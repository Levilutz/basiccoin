package main

import (
	"fmt"
	"net"
	"sync"

	"github.com/levilutz/basiccoin/src/mainbus"
	"github.com/levilutz/basiccoin/src/peer"
	"github.com/levilutz/basiccoin/src/util"
)

func addPeer(
	conn *net.TCPConn,
	peerBuses map[string]*peer.PeerBus,
	peerBusesMutex *sync.Mutex,
	mainBus *mainbus.MainBus,
) {
	bus, err := peer.ReceivePeerGreeting(peer.NewPeerConn(conn), mainBus)
	if err != nil {
		fmt.Println("Failed to establish with new peer:", err.Error())
		return
	}
	peerBusesMutex.Lock()
	defer peerBusesMutex.Unlock()
	if _, ok := peerBuses[bus.PeerRuntimeID]; !ok {
		peerBuses[bus.PeerRuntimeID] = bus
	} else {
		bus.UrgentEvents <- peer.PeerBusEvent{
			ShouldEnd: &peer.ShouldEndEvent{
				SendClose:     true,
				NotifyMainBus: false,
			},
		}
	}
}

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

	for {
		select {
		case conn := <-conns:
			go addPeer(conn, peerBuses, &peerBusesMutex, mainBus)
		case closedID := <-mainBus.PeerClosings:
			peerBusesMutex.Lock()
			delete(peerBuses, closedID)
			peerBusesMutex.Unlock()
		default:
			select {
			case conn := <-conns:
				go addPeer(conn, peerBuses, &peerBusesMutex, mainBus)
			case closedID := <-mainBus.PeerClosings:
				peerBusesMutex.Lock()
				delete(peerBuses, closedID)
				peerBusesMutex.Unlock()
			case event := <-mainBus.UrgentEvents:
				fmt.Println(event)
			default:
				select {
				case conn := <-conns:
					go addPeer(conn, peerBuses, &peerBusesMutex, mainBus)
				case closedID := <-mainBus.PeerClosings:
					peerBusesMutex.Lock()
					delete(peerBuses, closedID)
					peerBusesMutex.Unlock()
				case event := <-mainBus.UrgentEvents:
					fmt.Println(event)
				case event := <-mainBus.Events:
					fmt.Println(event)
				}
			}
		}
	}

	// TODO
	// Make main loop receive events from 1. listener channel and
	// 2. peer bus kill signals
	// 3. all peer buses (flattened?? and randomly shuffled per round)
	// On new conn, make bus and receive greeting
}
