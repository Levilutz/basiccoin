package main

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/levilutz/basiccoin/src/events"
	"github.com/levilutz/basiccoin/src/peer"
	"github.com/levilutz/basiccoin/src/util"
)

// TODO: Make funcs here methods, rename (Manager?), move to module
type MainState struct {
	newConnChannel   chan *net.TCPConn
	mainBus          chan events.MainEvent
	peers            map[string]*peer.Peer
	peersMu          sync.Mutex
	knownPeerAddrs   map[string]struct{}
	knownPeerAddrsMu sync.Mutex
}

func managerRoutine(state *MainState) {
	if util.Constants.DebugManagerLoop {
		fmt.Println("MANAGER_LOOP")
	}
	if util.Constants.DebugTicker {
		go blankTicker()
	}
	filterKnownPeersTicker := time.NewTicker(util.Constants.FilterKnownPeersFreq)
	for {
		select {
		case conn := <-state.newConnChannel:
			if util.Constants.DebugManagerLoop {
				fmt.Println("MANAGER_CONN", conn)
			}
			go addPeer(state, conn)

		case event := <-state.mainBus:
			if util.Constants.DebugManagerLoop {
				fmt.Println("MANAGER_BUS", event)
			}
			go handleMainBusEvent(state, event)

		case <-filterKnownPeersTicker.C:
			if util.Constants.DebugManagerLoop {
				fmt.Println("MANAGER_FILTER")
			}
			go filterKnownPeers(state)
		}
	}
}

func blankTicker() {
	ticker := time.NewTicker(time.Second)
	for {
		<-ticker.C
		fmt.Println(".")
	}
}

func addPeer(
	state *MainState,
	conn *net.TCPConn,
) {
	p, err := peer.NewPeerInbound(conn, state.mainBus)
	if err != nil {
		fmt.Println("Failed to establish with new peer:", err.Error())
		return
	}
	go p.Loop()
	state.peersMu.Lock()
	defer state.peersMu.Unlock()
	if _, ok := state.peers[p.HelloMsg.RuntimeID]; !ok {
		state.peers[p.HelloMsg.RuntimeID] = p
	} else {
		p.EventBus <- events.PeerEvent{
			ShouldEnd: &events.ShouldEndPeerEvent{
				NotifyMainBus: false,
			},
		}
	}
}

func handleMainBusEvent(state *MainState, event events.MainEvent) {
	if msg := event.PeerClosing; msg != nil {
		state.peersMu.Lock()
		delete(state.peers, msg.RuntimeID)
		state.peersMu.Unlock()

	} else if msg := event.PeersReceived; msg != nil {
		// TODO Verify this peer before insert (in goroutine)
		state.knownPeerAddrsMu.Lock()
		for _, addr := range msg.PeerAddrs {
			state.knownPeerAddrs[addr] = struct{}{}
		}
		state.knownPeerAddrsMu.Unlock()

	} else if msg := event.PeersWanted; msg != nil {
		state.peers[msg.PeerRuntimeID].EventBus <- events.PeerEvent{
			PeersData: &events.PeersDataPeerEvent{
				PeerAddrs: getKnownPeersList(state),
			},
		}

	} else {
		fmt.Println("Unhandled event", event)
	}
}

func filterKnownPeers(state *MainState) {
	fmt.Println("Known Peers:", getKnownPeersList(state))
}

func getKnownPeersList(state *MainState) []string {
	addrs := make([]string, 0)
	state.knownPeerAddrsMu.Lock()
	defer state.knownPeerAddrsMu.Unlock()
	for addr := range state.knownPeerAddrs {
		addrs = append(addrs, addr)
	}
	return addrs
}
