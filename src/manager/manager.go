package manager

import (
	"fmt"
	"math/rand"
	"net"
	"sync/atomic"
	"time"

	"github.com/levilutz/basiccoin/src/db"
	"github.com/levilutz/basiccoin/src/events"
	"github.com/levilutz/basiccoin/src/miner"
	"github.com/levilutz/basiccoin/src/peer"
	"github.com/levilutz/basiccoin/src/util"
)

type MetConn struct {
	PeerConn       *peer.PeerConn
	HelloMsg       *peer.HelloMessage
	WeAreInitiator bool
}

type Manager struct {
	metConnChannel chan MetConn
	mainBus        chan any
	peers          map[string]*peer.Peer
	inv            *db.Inv
	miners         []*miner.Miner
	minersAggCh    <-chan db.Block
	minersActive   atomic.Bool
}

func NewManager() *Manager {
	return &Manager{
		metConnChannel: make(chan MetConn),
		mainBus:        make(chan any),
		peers:          make(map[string]*peer.Peer),
		inv:            db.NewInv(),
		miners:         make([]*miner.Miner, util.Constants.Miners),
		minersAggCh:    nil,
		minersActive:   atomic.Bool{},
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

func (m *Manager) Mine() {
	if util.Constants.Miners == 0 {
		return
	}
	chs := make([]chan db.Block, util.Constants.Miners)
	for i := 0; i < util.Constants.Miners; i++ {
		m.miners[i] = miner.NewMiner()
		go m.miners[i].Loop()
		chs[i] = m.miners[i].SolutionCh
	}
	m.minersAggCh = util.Aggregate(chs)
	m.minersActive.Store(true)
}

func (m *Manager) Loop() {
	seekNewPeersTicker := time.NewTicker(util.Constants.SeekNewPeersFreq)
	printPeersUpdateTicker := time.NewTicker(util.Constants.PrintPeersUpdateFreq)
	for {
		select {
		case metConn := <-m.metConnChannel:
			m.addMetConn(metConn)

		case <-seekNewPeersTicker.C:
			m.seekNewPeers()

		case event := <-m.mainBus:
			m.handleMainBusEvent(event)

		case <-printPeersUpdateTicker.C:
			m.printPeersUpdate()

		case sol := <-m.minersAggCh:
			m.handleMinedSolution(sol)
		}
	}
}

func (m *Manager) addMetConn(metConn MetConn) {
	upgradeable := metConn.HelloMsg.RuntimeID != util.Constants.RuntimeID &&
		metConn.HelloMsg.Version == util.Constants.Version &&
		!m.peerConnected(metConn.HelloMsg.RuntimeID) &&
		len(m.peers) < util.Constants.MaxPeers

	if upgradeable {
		peer := peer.NewPeer(
			metConn.HelloMsg,
			metConn.PeerConn,
			m.mainBus,
			metConn.WeAreInitiator,
			m.inv,
		)
		go peer.Loop()
		m.peers[metConn.HelloMsg.RuntimeID] = peer

	} else {
		go func() {
			metConn.PeerConn.TransmitStringLine("cmd:close")
			metConn.PeerConn.C.Close()
		}()
	}
}

func (m *Manager) seekNewPeers() {
	if len(m.peers) >= util.Constants.MinPeers {
		return
	}
	peerInd := rand.Intn(len(m.peers))
	peerId := m.getPeerIDsList()[peerInd]
	go func() {
		m.peers[peerId].EventBus <- events.PeersWantedPeerEvent{}
	}()
}

func (m *Manager) getPeerIDsList() []string {
	ids := make([]string, len(m.peers))
	i := 0
	for addr := range m.peers {
		ids[i] = addr
		i++
	}
	return ids
}

func (m *Manager) getPeerAddrsList() []string {
	addrs := make([]string, 0)
	for _, peer := range m.peers {
		if peer.HelloMsg.Addr != "" {
			addrs = append(addrs, peer.HelloMsg.Addr)
		}
	}
	return addrs
}

func (m *Manager) peerConnected(runtimeID string) bool {
	_, exists := m.peers[runtimeID]
	return exists
}

func (m *Manager) handleMainBusEvent(event any) {
	switch msg := event.(type) {
	case events.PeerClosingMainEvent:
		delete(m.peers, msg.RuntimeID)

	case events.PeersReceivedMainEvent:
		if len(m.peers) >= util.Constants.MaxPeers {
			return
		}
		for _, addr := range msg.PeerAddrs {
			addr := addr
			go func() {
				pc, err := peer.ResolvePeerConn(addr)
				if err == nil {
					m.IntroducePeerConn(pc, true)
				}
			}()
		}

	case events.PeersWantedMainEvent:
		addrs := m.getPeerAddrsList()
		if len(addrs) == 0 {
			return
		}
		go func() {
			m.peers[msg.PeerRuntimeID].EventBus <- events.PeersDataPeerEvent{
				PeerAddrs: addrs,
			}
		}()

	case events.CandidateLedgerUpgradeMainEvent:
		// TODO: Verify synchronously
		// if failed, blacklist head id
		// if passed, loop StoreFullBlock from bottom of tree, then update head

	default:
		fmt.Printf("Unhandled main event %T\n", event)
	}
}

func (m *Manager) printPeersUpdate() {
	fmt.Println("peers:", len(m.peers), m.getPeerAddrsList())
}

func (m *Manager) IntroducePeerConn(pc *peer.PeerConn, weAreInitiator bool) {
	helloMsg := pc.Handshake()
	m.metConnChannel <- MetConn{
		PeerConn:       pc,
		HelloMsg:       helloMsg,
		WeAreInitiator: weAreInitiator,
	}
}

func (m *Manager) setMinersTarget(target db.Block) {
	// Wait until miners ready
	ready := false
	for i := 0; i < 10; i++ {
		ready = m.minersActive.Load()
		if ready {
			break
		}
		time.Sleep(time.Second)
	}
	if !ready {
		return
	}
	// Set each target
	for i := 0; i < util.Constants.Miners; i++ {
		m.miners[i].SetTarget(target)
	}
}

func (m *Manager) handleMinedSolution(sol db.Block) {
	// Verify solution
	// Insert solution
}
