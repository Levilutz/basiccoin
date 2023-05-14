package main

import (
	"fmt"
	"net"
	"sync"

	"github.com/levilutz/basiccoin/src/mainbus"
	"github.com/levilutz/basiccoin/src/peer"
)

type MainState struct {
	newConnChannel          chan *net.TCPConn
	mainBus                 *mainbus.MainBus
	peers                   map[string]*peer.Peer
	peersMutex              sync.Mutex
	candidatePeerAddrs      map[string]struct{}
	candidatePeerAddrsMutex sync.Mutex
}

func managerRoutine(state *MainState) {
	for {
		select {
		case conn := <-state.newConnChannel:
			go addPeer(state, conn)
		case event := <-state.mainBus.Events:
			go handleMainBusEvent(state, event)
		}
	}
}

func addPeer(
	state *MainState,
	conn *net.TCPConn,
) {
	p, err := peer.ReceivePeerGreeting(peer.NewPeerConn(conn), state.mainBus)
	if err != nil {
		fmt.Println("Failed to establish with new peer:", err.Error())
		return
	}
	state.peersMutex.Lock()
	defer state.peersMutex.Unlock()
	if _, ok := state.peers[p.HelloMsg.RuntimeID]; !ok {
		state.peers[p.HelloMsg.RuntimeID] = p
	} else {
		p.Events <- peer.PeerEvent{
			ShouldEnd: &peer.ShouldEndEvent{
				SendClose:     true,
				NotifyMainBus: false,
			},
		}
	}
}

func handleMainBusEvent(state *MainState, event mainbus.MainBusEvent) {
	if msg := event.PeerClosing; msg != nil {
		state.peersMutex.Lock()
		delete(state.peers, msg.RuntimeID)
		state.peersMutex.Unlock()

	} else if msg := event.PeersReceived; msg != nil {
		state.candidatePeerAddrsMutex.Lock()
		for _, addr := range msg.PeerAddrs {
			state.candidatePeerAddrs[addr] = struct{}{}
		}
		state.candidatePeerAddrsMutex.Unlock()

	} else if msg := event.PeersWanted; msg != nil {

	} else {
		fmt.Println("Unhandled event", event)
	}
}
