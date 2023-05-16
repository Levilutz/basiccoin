package manager

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/levilutz/basiccoin/src/events"
	"github.com/levilutz/basiccoin/src/peer"
	"github.com/levilutz/basiccoin/src/util"
)

type Manager struct {
	newConnChannel   chan *net.TCPConn
	mainBus          chan events.MainEvent
	peers            map[string]*peer.Peer
	peersMu          sync.Mutex
	knownPeerAddrs   map[string]struct{}
	knownPeerAddrsMu sync.Mutex
}

func NewManager(
	newConnChannel chan *net.TCPConn,
	mainBus chan events.MainEvent,
	peers map[string]*peer.Peer,
	knownPeerAddrs map[string]struct{},
) *Manager {
	return &Manager{
		newConnChannel: newConnChannel,
		mainBus:        mainBus,
		peers:          peers,
		knownPeerAddrs: knownPeerAddrs,
	}
}

func (m *Manager) Loop() {
	if util.Constants.DebugManagerLoop {
		fmt.Println("MANAGER_LOOP")
	}
	if util.Constants.DebugTicker {
		go blankTicker()
	}
	filterKnownPeersTicker := time.NewTicker(util.Constants.FilterKnownPeersFreq)
	for {
		select {
		case conn := <-m.newConnChannel:
			if util.Constants.DebugManagerLoop {
				fmt.Println("MANAGER_CONN", conn)
			}
			go m.addInboundPeer(conn)

		case event := <-m.mainBus:
			if util.Constants.DebugManagerLoop {
				fmt.Println("MANAGER_BUS", event)
			}
			go m.handleMainBusEvent(event)

		case <-filterKnownPeersTicker.C:
			if util.Constants.DebugManagerLoop {
				fmt.Println("MANAGER_FILTER")
			}
			go m.filterKnownPeers()
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

func (m *Manager) addInboundPeer(conn *net.TCPConn) {
	p, err := peer.NewPeerInbound(conn, m.mainBus)
	if err != nil {
		fmt.Println("failed to establish with new peer:", err.Error())
		return
	}
	go p.Loop()
	m.peersMu.Lock()
	defer m.peersMu.Unlock()
	if _, ok := m.peers[p.HelloMsg.RuntimeID]; !ok {
		m.peers[p.HelloMsg.RuntimeID] = p
	} else {
		p.EventBus <- events.PeerEvent{
			ShouldEnd: &events.ShouldEndPeerEvent{
				NotifyMainBus: false,
			},
		}
	}
}

func (m *Manager) handleMainBusEvent(event events.MainEvent) {
	if msg := event.PeerClosing; msg != nil {
		m.peersMu.Lock()
		delete(m.peers, msg.RuntimeID)
		m.peersMu.Unlock()

	} else if msg := event.PeersReceived; msg != nil {
		// TODO Verify this peer before insert (in goroutine)
		m.knownPeerAddrsMu.Lock()
		for _, addr := range msg.PeerAddrs {
			m.knownPeerAddrs[addr] = struct{}{}
		}
		m.knownPeerAddrsMu.Unlock()

	} else if msg := event.PeersWanted; msg != nil {
		m.peers[msg.PeerRuntimeID].EventBus <- events.PeerEvent{
			PeersData: &events.PeersDataPeerEvent{
				PeerAddrs: m.getKnownPeersList(),
			},
		}

	} else {
		fmt.Println("Unhandled event", event)
	}
}

func (m *Manager) filterKnownPeers() {
	fmt.Println("known Peers:", m.getKnownPeersList())
}

func (m *Manager) getKnownPeersList() []string {
	addrs := make([]string, 0)
	m.knownPeerAddrsMu.Lock()
	defer m.knownPeerAddrsMu.Unlock()
	for addr := range m.knownPeerAddrs {
		addrs = append(addrs, addr)
	}
	return addrs
}