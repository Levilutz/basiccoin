package peerfactory

import (
	"math/rand"
	"net"
	"time"

	"github.com/levilutz/basiccoin/internal/peer"
	"github.com/levilutz/basiccoin/internal/pubsub"
	"github.com/levilutz/basiccoin/pkg/prot"
	"github.com/levilutz/basiccoin/pkg/set"
	"github.com/levilutz/basiccoin/pkg/syncqueue"
	"github.com/levilutz/basiccoin/pkg/topic"
	"github.com/levilutz/basiccoin/pkg/util"
)

// The peer factory's subscriptions.
// Ensure each of these is initialized in NewPeerFactory.
type subcriptions struct {
	PeerAnnouncedAddr *topic.SubCh[pubsub.PeerAnnouncedAddrEvent]
	PeerClosing       *topic.SubCh[pubsub.PeerClosingEvent]
	PeersReceived     *topic.SubCh[pubsub.PeersReceivedEvent]
	PeersRequested    *topic.SubCh[pubsub.PeersRequestedEvent]
}

// A peer factory. Does not manage the peers after creation.
// May listen for inbound connections and/or seek new outbound connections.
// Keeps track of what peers exist.
type PeerFactory struct {
	params         Params
	pubSub         *pubsub.PubSub
	subs           *subcriptions
	newConns       chan *prot.Conn
	newAddrs       *syncqueue.SyncQueue[string]
	knownPeers     *set.Set[string]
	knownPeerAddrs map[string]string // Not all knownPeers appear here
}

// Create a new peer factory given a message bus instance.
func NewPeerFactory(params Params, pubSub *pubsub.PubSub) *PeerFactory {
	subs := &subcriptions{
		PeerAnnouncedAddr: pubSub.PeerAnnouncedAddr.SubCh(),
		PeerClosing:       pubSub.PeerClosing.SubCh(),
		PeersReceived:     pubSub.PeersReceived.SubCh(),
		PeersRequested:    pubSub.PeersRequested.SubCh(),
	}
	return &PeerFactory{
		params:         params,
		pubSub:         pubSub,
		subs:           subs,
		newConns:       make(chan *prot.Conn, 256),
		newAddrs:       syncqueue.NewSyncQueue[string](),
		knownPeers:     set.NewSet[string](),
		knownPeerAddrs: make(map[string]string),
	}
}

// Start the peer factory's loop.
func (pf *PeerFactory) Loop() {
	go pf.tryNewAddrs()

	// Start listener if desired
	if pf.params.Listen {
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

		case event := <-pf.subs.PeersReceived.C:
			for runtimeId, addr := range event.PeerAddrs {
				if !pf.knownPeers.Includes(runtimeId) {
					pf.newAddrs.Push(addr)
				}
			}

		case event := <-pf.subs.PeersRequested.C:
			pf.pubSub.SendPeers.Pub(pubsub.SendPeersEvent{
				TargetRuntimeId: event.PeerRuntimeId,
				PeerAddrs:       util.CopyMap(pf.knownPeerAddrs),
			})

		case <-seekPeersTicker.C:
			pf.seekNewPeers()
		}
	}
}

// Receive and attempt to connect to new addrs.
func (pf *PeerFactory) tryNewAddrs() {
	for {
		for addr, ok := pf.newAddrs.Pop(); ok; addr, ok = pf.newAddrs.Pop() {
			protParams := prot.NewParams(pf.params.RuntimeId, true)
			conn, err := prot.ResolveConn(protParams, addr)
			if err != nil {
				continue
			} else if conn.HasErr() {
				conn.CloseIfPossible()
				continue
			}
			pf.newConns <- conn

		}
		time.Sleep(time.Millisecond * 25)
	}
}

// Routine to start listening for new connections.
func (pf *PeerFactory) listen() {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:21720")
	if err != nil {
		panic(err)
	}
	listen, err := net.ListenTCP("tcp", addr)
	if err != nil {
		panic(err)
	}
	defer listen.Close()
	for {
		tcpConn, err := listen.AcceptTCP()
		if err != nil {
			continue
		}
		protParams := prot.NewParams(pf.params.RuntimeId, false)
		conn := prot.NewConn(protParams, tcpConn)
		if conn.HasErr() {
			conn.CloseIfPossible()
			continue
		}
		pf.newConns <- conn
	}
}

// Upgrade a connection to peer, if appropriate.
func (pf *PeerFactory) addConn(conn *prot.Conn) {
	if conn.HasErr() {
		conn.CloseIfPossible()
		return
	}
	if pf.knownPeers.Size() < pf.params.MaxPeers ||
		pf.knownPeers.Includes(conn.PeerRuntimeId()) {
		// Upgrade to peer
		go peer.NewPeer(pf.pubSub, conn).Loop()
		pf.knownPeers.Add(conn.PeerRuntimeId())
	} else {
		// Try to inform them we're closing, ignore any errs
		conn.CloseIfPossible()
	}
}

// Check if we should and can seek new peers, then do so.
func (pf *PeerFactory) seekNewPeers() {
	if pf.knownPeers.Size() == 0 || pf.knownPeers.Size() >= pf.params.MinPeers {
		return
	}
	targetInd := rand.Intn(pf.knownPeers.Size())
	targetRuntimeId := pf.knownPeers.ToList()[targetInd]
	pf.pubSub.ShouldRequestPeers.Pub(pubsub.ShouldRequestPeersEvent{
		TargetRuntimeId: targetRuntimeId,
	})
}
