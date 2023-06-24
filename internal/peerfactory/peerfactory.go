package peerfactory

import (
	"fmt"
	"math/rand"
	"net"
	"sync/atomic"
	"time"

	"github.com/levilutz/basiccoin/internal/bus"
	"github.com/levilutz/basiccoin/internal/inv"
	"github.com/levilutz/basiccoin/internal/peer"
	"github.com/levilutz/basiccoin/pkg/core"
	"github.com/levilutz/basiccoin/pkg/prot"
	"github.com/levilutz/basiccoin/pkg/set"
	"github.com/levilutz/basiccoin/pkg/syncqueue"
	"github.com/levilutz/basiccoin/pkg/topic"
	"github.com/levilutz/basiccoin/pkg/util"
)

// The peer factory's subscriptions.
// Ensure each of these is initialized in NewPeerFactory.
type subcriptions struct {
	PeerAnnouncedAddr *topic.SubCh[bus.PeerAnnouncedAddrEvent]
	PeerClosing       *topic.SubCh[bus.PeerClosingEvent]
	PeersReceived     *topic.SubCh[bus.PeersReceivedEvent]
	PeersRequested    *topic.SubCh[bus.PeersRequestedEvent]
	PrintUpdate       *topic.SubCh[bus.PrintUpdateEvent]
	ValidatedHead     *topic.SubCh[bus.ValidatedHeadEvent]
}

// A peer factory. Does not manage the peers after creation.
// May listen for inbound connections and/or seek new outbound connections.
// Keeps track of what peers exist.
type PeerFactory struct {
	params         Params
	bus            *bus.Bus
	inv            inv.InvReader
	subs           *subcriptions
	newConns       chan *prot.Conn
	newAddrs       *syncqueue.SyncQueue[string]
	knownPeers     *set.Set[string]
	knownPeerAddrs map[string]string // Not all knownPeers appear here
	listenStarted  atomic.Bool
	seedAddrs      []string
	curHead        core.HashT
}

// Create a new peer factory given a message bus instance.
func NewPeerFactory(params Params, msgBus *bus.Bus, inv inv.InvReader) *PeerFactory {
	subs := &subcriptions{
		PeerAnnouncedAddr: msgBus.PeerAnnouncedAddr.SubCh(),
		PeerClosing:       msgBus.PeerClosing.SubCh(),
		PeersReceived:     msgBus.PeersReceived.SubCh(),
		PeersRequested:    msgBus.PeersRequested.SubCh(),
		PrintUpdate:       msgBus.PrintUpdate.SubCh(),
		ValidatedHead:     msgBus.ValidatedHead.SubCh(),
	}
	return &PeerFactory{
		params:         params,
		bus:            msgBus,
		inv:            inv,
		subs:           subs,
		newConns:       make(chan *prot.Conn, 256),
		newAddrs:       syncqueue.NewSyncQueue[string](),
		knownPeers:     set.NewSet[string](),
		knownPeerAddrs: make(map[string]string),
		seedAddrs:      make([]string, 0),
		curHead:        core.HashT{},
	}
}

// Set our seed peer. Must run before Loop.
func (pf *PeerFactory) SetSeeds(seedAddrs []string) {
	pf.seedAddrs = seedAddrs
}

// Start the peer factory's loop.
func (pf *PeerFactory) Loop() {
	go pf.tryNewAddrs()

	// Try alternating connections to each seed peer until we get success
	if len(pf.seedAddrs) > 0 {
		// Try to connect to any of the seed peers
		numTries := 15
		for i := 0; i < numTries; i++ {
			found := false
			for _, seedAddr := range pf.seedAddrs {
				conn, err := pf.tryConn(seedAddr)
				if err == nil {
					pf.newConns <- conn
					found = true
					fmt.Println("successfully connected to seed peer")
					break
				} else {
					fmt.Printf("failed to connect to seed peer: %s\n", err.Error())
				}
			}
			if found {
				break
			}
			if i == numTries-1 {
				panic(fmt.Sprintf("failed to reach seed peer after %d tries", numTries))
			}
			time.Sleep(time.Second)
		}
		// Queue the addrs so we can connect to the other seeds anyway
		pf.newAddrs.Push(pf.seedAddrs...)
	}

	// Start listener if desired
	if pf.params.Listen && pf.params.LocalAddr != "" {
		go pf.listen()
	}

	// Loop
	seekPeersTicker := time.NewTicker(pf.params.SeekNewPeersFreq)
	for {
		select {
		case conn := <-pf.newConns:
			pf.addConn(conn)

		case event := <-pf.subs.PeerAnnouncedAddr.C:
			pf.knownPeerAddrs[event.PeerRuntimeId] = event.Addr

		case event := <-pf.subs.PeerClosing.C:
			pf.knownPeers.Remove(event.PeerRuntimeId)
			delete(pf.knownPeerAddrs, event.PeerRuntimeId)

		case event := <-pf.subs.PeersReceived.C:
			for runtimeId, addr := range event.PeerAddrs {
				if runtimeId != pf.params.RuntimeId && !pf.knownPeers.Includes(runtimeId) {
					pf.newAddrs.Push(addr)
				}
			}

		case event := <-pf.subs.PeersRequested.C:
			pf.bus.SendPeers.Pub(bus.SendPeersEvent{
				TargetRuntimeId: event.PeerRuntimeId,
				PeerAddrs:       util.CopyMap(pf.knownPeerAddrs),
			})

		case event := <-pf.subs.ValidatedHead.C:
			pf.curHead = event.Head

		case event := <-pf.subs.PrintUpdate.C:
			if !event.PeerFactory {
				continue
			}
			fmt.Printf("peers: %d\n", pf.knownPeers.Size())

		case <-seekPeersTicker.C:
			pf.seekNewPeers()
		}
	}
}

// Receive and attempt to connect to new addrs.
func (pf *PeerFactory) tryNewAddrs() {
	for {
		for addr, ok := pf.newAddrs.Pop(); ok; addr, ok = pf.newAddrs.Pop() {
			conn, err := pf.tryConn(addr)
			if err != nil {
				fmt.Printf("failed to resolve addr %s: %s\n", addr, err.Error())
				continue
			}
			pf.newConns <- conn
		}
		time.Sleep(time.Millisecond * 25)
	}
}

// Try to connect to the given addr.
func (pf *PeerFactory) tryConn(addr string) (*prot.Conn, error) {
	protParams := prot.NewParams(pf.params.RuntimeId, true, pf.params.DebugConns)
	conn, err := prot.ResolveConn(protParams, addr)
	if err != nil {
		return nil, err
	} else if conn.HasErr() {
		conn.CloseIfPossible(nil)
		return nil, conn.Err()
	}
	return conn, nil
}

// Routine to start listening for new connections.
func (pf *PeerFactory) listen() {
	// Guard against multiple simultaneous litens. This isn't perfect but it's just-in-case.
	if pf.listenStarted.Load() {
		return
	}
	pf.listenStarted.Store(true)

	// Get the listener
	if pf.params.LocalAddr == "" {
		panic("cannot listen without local addr set")
	}
	addr, err := net.ResolveTCPAddr("tcp", pf.params.LocalAddr)
	if err != nil {
		panic(err)
	}
	listen, err := net.ListenTCP("tcp", addr)
	if err != nil {
		panic(err)
	}
	defer listen.Close()

	// Loop accepting new connections
	for {
		tcpConn, err := listen.AcceptTCP()
		if err != nil {
			continue
		}
		protParams := prot.NewParams(pf.params.RuntimeId, false, pf.params.DebugConns)
		conn := prot.NewConn(protParams, tcpConn)
		if conn.HasErr() {
			conn.CloseIfPossible(nil)
			continue
		}
		pf.newConns <- conn
	}
}

// Upgrade a connection to peer, if appropriate.
func (pf *PeerFactory) addConn(conn *prot.Conn) {
	if conn.HasErr() {
		conn.CloseIfPossible(nil)
		return
	}
	runtimeId := conn.PeerRuntimeId()
	if pf.knownPeers.Size() < pf.params.MaxPeers &&
		!pf.knownPeers.Includes(runtimeId) {
		// Upgrade to peer
		go peer.NewPeer(pf.bus, pf.inv, conn, pf.curHead).Loop()
		pf.knownPeers.Add(runtimeId)
		// Set our localaddr and start listen if we only now can
		if pf.params.Listen && pf.params.LocalAddr == "" {
			pf.params.LocalAddr = conn.LocalAddr().IP.String() + ":21720"
			go pf.listen()
		}
		// Broadcast our localaddr to the peer if we want to listen
		if pf.params.Listen {
			pf.bus.ShouldAnnounceAddr.Pub(bus.ShouldAnnounceAddrEvent{
				TargetRuntimeId: runtimeId,
				Addr:            pf.params.LocalAddr,
			})
		}
	} else {
		fmt.Printf("will not connect to peer %s\n", conn.PeerRuntimeId())
		// Try to inform them we're closing, ignore any errs
		conn.CloseIfPossible(util.CopyMap(pf.knownPeerAddrs))
	}
}

// Check if we should and can seek new peers, then do so.
func (pf *PeerFactory) seekNewPeers() {
	if pf.knownPeers.Size() == 0 || pf.knownPeers.Size() >= pf.params.MinPeers {
		return
	}
	targetInd := rand.Intn(pf.knownPeers.Size())
	targetRuntimeId := pf.knownPeers.ToList()[targetInd]
	pf.bus.ShouldRequestPeers.Pub(bus.ShouldRequestPeersEvent{
		TargetRuntimeId: targetRuntimeId,
	})
}
