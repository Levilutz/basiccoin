package peerfactory

import (
	"math/rand"
	"net"
	"time"

	"github.com/levilutz/basiccoin/internal/peer"
	"github.com/levilutz/basiccoin/internal/pubsub"
	"github.com/levilutz/basiccoin/pkg/prot"
	"github.com/levilutz/basiccoin/pkg/set"
	"github.com/levilutz/basiccoin/pkg/topic"
)

// The peer factory's subscriptions.
type subcriptions struct {
	PeerClosing *topic.SubCh[pubsub.PeerClosingEvent]
}

// A peer factory. Does not manage the peers after creation.
// May listen for inbound connections and/or seek new outbound connections.
// Keeps track of what peers exist.
type PeerFactory struct {
	params     Params
	pubSub     *pubsub.PubSub
	subs       *subcriptions
	newConns   chan *prot.Conn
	knownPeers set.Set[string] // Not sync-safe, should only access from main routine
}

// Create a new peer factory given a message bus instance.
func NewPeerFactory(params Params, pubSub *pubsub.PubSub) *PeerFactory {
	subs := &subcriptions{
		PeerClosing: pubSub.PeerClosing.SubCh(),
	}
	return &PeerFactory{
		params:   params,
		pubSub:   pubSub,
		subs:     subs,
		newConns: make(chan *prot.Conn, 256),
	}
}

// Start the peer factory's loop.
func (pf *PeerFactory) Loop() {
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
		case peerClosingEvent := <-pf.subs.PeerClosing.C:
			pf.knownPeers.Remove(peerClosingEvent.PeerRuntimeId)
		case <-seekPeersTicker.C:
			pf.seekNewPeers()
		}
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
			continue
		}
		pf.newConns <- conn
	}
}

// Upgrade a connection to peer, if appropriate.
func (pf *PeerFactory) addConn(conn *prot.Conn) {
	if conn.HasErr() {
		return
	}
	if pf.knownPeers.Size() < pf.params.MaxPeers ||
		pf.knownPeers.Includes(conn.PeerRuntimeId()) {
		// Upgrade to peer
		go peer.NewPeer(pf.pubSub, conn).Loop()
		pf.knownPeers.Add(conn.PeerRuntimeId())
	} else {
		// Try to inform them we're closing, ignore any errs
		go func() {
			defer func() {
				recover()
			}()
			conn.WriteString("cmd:close")
			conn.Close()
		}()
	}
}

// Check if we should and can seek new peers, then do so.
func (pf *PeerFactory) seekNewPeers() {
	if pf.knownPeers.Size() == 0 || pf.knownPeers.Size() >= pf.params.MinPeers {
		return
	}
	// TODO: Request a random peer to send over their peers
	peerInd := rand.Intn(pf.knownPeers.Size())
	peerRuntimeId := pf.knownPeers.ToList()[peerInd]
	pf.pubSub.ShouldRequestPeers.Pub(pubsub.ShouldRequestPeersEvent{
		PeerRuntimeId: peerRuntimeId,
	})
}
