package main

import (
	"fmt"
	"net"
	"sync"

	"github.com/levilutz/basiccoin/src/mainbus"
	"github.com/levilutz/basiccoin/src/peer"
)

type MainState struct {
	newConnChannel chan *net.TCPConn
	mainBus        *mainbus.MainBus
	peerBuses      map[string]*peer.PeerBus
	peerBusesMutex *sync.Mutex
}

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
		bus.Events <- peer.PeerBusEvent{
			ShouldEnd: &peer.ShouldEndEvent{
				SendClose:     true,
				NotifyMainBus: false,
			},
		}
	}
}

func managerRoutine(state MainState) {
	for {
		select {
		case conn := <-state.newConnChannel:
			go addPeer(conn, state.peerBuses, state.peerBusesMutex, state.mainBus)
		case event := <-state.mainBus.Events:
			go handleMainBusEvent(state, event)
		}
	}
}

func handleMainBusEvent(state MainState, event mainbus.MainBusEvent) {
	if msg := event.PeerClosing; msg != nil {
		state.peerBusesMutex.Lock()
		delete(state.peerBuses, msg.RuntimeID)
		state.peerBusesMutex.Unlock()
	} else {
		fmt.Println("Unrecognized event", msg)
	}
}
