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

type MetConn struct {
	PeerConn       *peer.PeerConn
	HelloMsg       *peer.HelloMessage
	WeAreInitiator bool
}

type Manager struct {
	metConnChannel   chan MetConn
	mainBus          chan events.MainEvent
	peers            map[string]*peer.Peer
	peersMu          sync.Mutex
	knownPeerAddrs   map[string]struct{}
	knownPeerAddrsMu sync.Mutex
}

func NewManager() *Manager {
	return &Manager{
		metConnChannel: make(chan MetConn),
		mainBus:        make(chan events.MainEvent),
		peers:          make(map[string]*peer.Peer),
		knownPeerAddrs: make(map[string]struct{}, 0),
	}
}

func (m *Manager) Listen() {
	addr, err := net.ResolveTCPAddr("tcp", util.Constants.LocalAddr)
	util.PanicErr(err)
	listen, err := net.ListenTCP("tcp", addr)
	util.PanicErr(err)
	defer listen.Close()
	for {
		conn, err := listen.AcceptTCP()
		if err != nil {
			continue
		}
		pc := peer.NewPeerConn(conn)
		helloMsg := pc.Handshake()
		m.metConnChannel <- MetConn{
			PeerConn:       pc,
			HelloMsg:       helloMsg,
			WeAreInitiator: false,
		}
	}
}

func (m *Manager) Loop() {
	filterKnownPeersTicker := time.NewTicker(util.Constants.FilterKnownPeersFreq)
	printPeersUpdateTicker := time.NewTicker(util.Constants.PrintPeersUpdateFreq)
	for {
		select {
		case metConn := <-m.metConnChannel:
			m.addMetConn(metConn)

		case event := <-m.mainBus:
			go m.handleMainBusEvent(event)

		case <-filterKnownPeersTicker.C:
			go m.filterKnownPeers()

		case <-printPeersUpdateTicker.C:
			go m.printPeersUpdate()
		}
	}
}

func (m *Manager) addMetConn(metConn MetConn) {
	m.peersMu.Lock()
	defer m.peersMu.Unlock()
	upgradeable := metConn.HelloMsg.RuntimeID != util.Constants.RuntimeID &&
		metConn.HelloMsg.Version == util.Constants.Version &&
		!m.peerConnected(metConn.HelloMsg.RuntimeID)

	if upgradeable {
		peer := peer.NewPeer(
			metConn.HelloMsg, metConn.PeerConn, m.mainBus, metConn.WeAreInitiator,
		)
		go peer.Loop()
		m.peers[metConn.HelloMsg.RuntimeID] = peer

	} else {
		metConn.PeerConn.TransmitStringLine("cmd:close")
		metConn.PeerConn.C.Close()
	}
}

// Must be called from locked context
func (m *Manager) peerConnected(runtimeID string) bool {
	_, exists := m.peers[runtimeID]
	return exists
}

func (m *Manager) handleMainBusEvent(event events.MainEvent) {
	if msg := event.PeerClosing; msg != nil {
		m.peersMu.Lock()
		delete(m.peers, msg.RuntimeID)
		m.peersMu.Unlock()

	} else if msg := event.PeersReceived; msg != nil {
		// TODO Verify this peer before insert (in goroutine)
		for _, addr := range msg.PeerAddrs {
			addr := addr
			go func() {
				pc, err := peer.ResolvePeerConn(addr)
				if err == nil {
					pc.VerifyAndClose()
					err = pc.Err()
				}
				if err == nil {
					m.knownPeerAddrsMu.Lock()
					m.knownPeerAddrs[addr] = struct{}{}
					m.knownPeerAddrsMu.Unlock()
				}
			}()
		}

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
	knownPeerAddrs := m.getKnownPeersList()
	for _, addr := range knownPeerAddrs {
		addr := addr
		go func() {
			pc, err := peer.ResolvePeerConn(addr)
			if err == nil {
				pc.VerifyAndClose()
				err = pc.Err()
			}
			if err != nil {
				m.knownPeerAddrsMu.Lock()
				delete(m.knownPeerAddrs, addr)
				m.knownPeerAddrsMu.Unlock()
			}
		}()
	}
}

func (m *Manager) printPeersUpdate() {
	m.knownPeerAddrsMu.Lock()
	defer m.knownPeerAddrsMu.Unlock()
	m.peersMu.Lock()
	defer m.peersMu.Unlock()
	fmt.Printf("peers:\t%d current,\t%dknown\n", len(m.peers), len(m.knownPeerAddrs))
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

func (m *Manager) IntroducePeerConn(pc *peer.PeerConn, weAreInitiator bool) {
	helloMsg := pc.Handshake()
	m.metConnChannel <- MetConn{
		PeerConn:       pc,
		HelloMsg:       helloMsg,
		WeAreInitiator: weAreInitiator,
	}
}
