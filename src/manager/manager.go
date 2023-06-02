package manager

import (
	"fmt"
	"math/rand"
	"net"
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
	state          *db.State
	minerSet       *miner.MinerSet
}

func NewManager() *Manager {
	inv := db.NewInv()
	state := db.NewState(inv)
	minerSet := miner.StartMinerSet(util.Constants.Miners)
	initialTarget := CreateMiningTarget(state, inv, db.HashTZero)
	minerSet.SetTargets(initialTarget)
	return &Manager{
		metConnChannel: make(chan MetConn),
		mainBus:        make(chan any),
		peers:          make(map[string]*peer.Peer),
		inv:            inv,
		state:          state,
		minerSet:       minerSet,
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

		case sol := <-m.minerSet.SolutionCh:
			solId := sol.Hash()
			fmt.Printf("mined potential solution: %x\n", solId)
			err := m.handleNewBestChain(
				solId, []db.Block{sol}, []db.MerkleNode{}, []db.Tx{},
			)
			if err != nil {
				fmt.Println("handleMinedSolution err:", err.Error())
			}
		}
	}
}

func (m *Manager) addMetConn(metConn MetConn) {
	if metConn.PeerConn.HasErr() {
		fmt.Println("unhandled pre-insertion peer err", metConn.PeerConn.Err().Error())
	}
	upgradeable := metConn.HelloMsg.RuntimeID != util.Constants.RuntimeID &&
		metConn.HelloMsg.Version == util.Constants.Version &&
		!m.peerConnected(metConn.HelloMsg.RuntimeID) &&
		len(m.peers) < util.Constants.MaxPeers

	if upgradeable {
		fmt.Printf("adding new peer: %s\n", metConn.HelloMsg.RuntimeID)
		peer := peer.NewPeer(
			metConn.HelloMsg,
			metConn.PeerConn,
			m.mainBus,
			metConn.WeAreInitiator,
			m.inv,
			m.state.GetHead(),
		)
		go peer.Loop()
		m.peers[metConn.HelloMsg.RuntimeID] = peer

	} else {
		if util.Constants.Debug {
			fmt.Printf("cancelling new peer: %s\n", metConn.HelloMsg.RuntimeID)
		}
		go func() {
			metConn.PeerConn.TransmitStringLine("cmd:close")
			metConn.PeerConn.Close()
		}()
	}
}

func (m *Manager) seekNewPeers() {
	if len(m.peers) == 0 || len(m.peers) >= util.Constants.MinPeers {
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

	case events.InboundSyncMainEvent:
		err := m.handleNewBestChain(msg.Head, msg.Blocks, msg.Merkles, msg.Txs)
		if err != nil {
			fmt.Println("failed to verify new chain:", err)
		}

	default:
		fmt.Printf("unhandled main event %T\n", event)
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

// Upgrades our chain to the given new head, if it proves to be better.
// Provide any blocks, merkles, or txs we might not know about (in the order to insert).
func (m *Manager) handleNewBestChain(
	newHead db.HashT,
	blocks []db.Block,
	merkles []db.MerkleNode,
	txs []db.Tx,
) error {
	oldHead := m.state.GetHead()
	// Insert each entity into the inventory, in order.
	for _, tx := range txs {
		if !m.inv.HasTx(tx.Hash()) {
			err := m.inv.StoreTx(tx)
			if err != nil {
				return err
			}
		}
	}
	for _, merkle := range merkles {
		if !m.inv.HasMerkle(merkle.Hash()) {
			err := m.inv.StoreMerkle(merkle)
			if err != nil {
				return err
			}
		}
	}
	for _, block := range blocks {
		if !m.inv.HasBlock(block.Hash()) {
			err := m.inv.StoreBlock(block)
			if err != nil {
				return err
			}
		}
	}
	// Verify new total work is higher
	if !m.inv.HasBlock(newHead) {
		return fmt.Errorf("provided head not known and not provided")
	}
	newWork := m.inv.GetBlockTotalWork(newHead)
	oldWork := m.inv.GetBlockTotalWork(oldHead)
	if !db.HashLT(oldWork, newWork) {
		return fmt.Errorf("new chain is not higher total work than current chain")
	}
	// Find common ancestor of our chain heads
	lcaId := m.inv.GetBlockLCA(oldHead, newHead)
	// Copy state, rewind to lca, and advance to new head
	newState := m.state.Copy()
	newState.RewindUntil(lcaId)
	newBlocks := m.inv.GetBlockAncestorsUntil(newHead, lcaId)
	// Advance through intermediate blocks, then the new head
	for i := len(newBlocks) - 1; i >= 0; i-- {
		if err := newState.Advance(newBlocks[i]); err != nil {
			return fmt.Errorf("failed to advance to mined block: %s", err.Error())
		}
	}
	if err := newState.Advance(newHead); err != nil {
		return fmt.Errorf("failed to advance to mined block: %s", err.Error())
	}
	// Shift to new head - this func shouldn't return err after this point
	fmt.Printf("upgrading head to %x - proven work %x\n", newState.GetHead(), newWork)
	m.state = newState
	// Set new miner targets
	target := CreateMiningTarget(m.state, m.inv, db.HashTZero)
	m.minerSet.SetTargets(target)
	// Broadcast solution to peers
	for _, p := range m.peers {
		p.SyncHead(newHead)
	}
	return nil
}
