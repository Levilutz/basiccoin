package peerfactory

import (
	"fmt"
	"net"
	"time"

	"github.com/levilutz/basiccoin/internal/pubsub"
	"github.com/levilutz/basiccoin/pkg/prot"
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
	params Params
	pubSub *pubsub.PubSub
	subs   *subcriptions
}

// Create a new peer factory given a message bus instance.
func NewPeerFactory(params Params, pubSub *pubsub.PubSub) *PeerFactory {
	subs := &subcriptions{
		PeerClosing: pubSub.PeerClosing.SubCh(),
	}
	return &PeerFactory{
		params: params,
		pubSub: pubSub,
		subs:   subs,
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
		case peerClosingEvent := <-pf.subs.PeerClosing.C:
			fmt.Println("peer closing received:", peerClosingEvent.PeerRuntimeId)
		case <-seekPeersTicker.C:
			fmt.Println("check if we need new peers")
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
		// Push conn to channel
	}
}
