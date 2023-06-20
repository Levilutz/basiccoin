package peer

import (
	"bytes"
	"fmt"
	"time"

	"github.com/levilutz/basiccoin/internal/pubsub"
	"github.com/levilutz/basiccoin/pkg/prot"
	"github.com/levilutz/basiccoin/pkg/topic"
)

// The peer's subscriptions.
type subscriptions struct {
	ShouldRequestPeers *topic.SubCh[pubsub.ShouldRequestPeersEvent]
	ValidatedHead      *topic.SubCh[pubsub.ValidatedHeadEvent]
}

// Close our subscriptions as we close.
func (s subscriptions) Close() {
	s.ValidatedHead.Close()
}

// A connection to a single peer.
type Peer struct {
	pubSub      *pubsub.PubSub
	subs        *subscriptions
	conn        *prot.Conn
	shouldClose bool
}

// Create a new peer given a message bus instance.
func NewPeer(pubSub *pubsub.PubSub, conn *prot.Conn) *Peer {
	subs := &subscriptions{
		ValidatedHead: pubSub.ValidatedHead.SubCh(),
	}
	return &Peer{
		pubSub: pubSub,
		subs:   subs,
		conn:   conn,
	}
}

// Start the peer's loop.
func (p *Peer) Loop() {
	// Handle panics and unsubscribe
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("peer closed from panic:", r)
		}
		p.pubSub.PeerClosing.Pub(pubsub.PeerClosingEvent{
			PeerRuntimeId: p.conn.PeerRuntimeId(),
		})
	}()

	// Loop
	for {
		select {
		case shouldRequestPeersEvent := <-p.subs.ShouldRequestPeers.C:
			if shouldRequestPeersEvent.PeerRuntimeId != p.conn.PeerRuntimeId() {
				continue
			}
			fmt.Println("should request their peers")

		case validatedHeadEvent := <-p.subs.ValidatedHead.C:
			fmt.Println("new validated head:", validatedHeadEvent.Head)

		default:
			msg := p.conn.ReadTimeout(time.Millisecond * 100)
			if p.conn.TimeoutErrOrPanic() != nil {
				continue
			}
			if err := p.handleReceivedMessage(msg); err != nil {
				fmt.Printf("error handling '%s': %s\n", msg, err.Error())
			}
		}
		if p.shouldClose {
			return
		}
	}
}

// Handle a message received from a peer.
func (p *Peer) handleReceivedMessage(msg []byte) error {
	if !bytes.HasPrefix(msg, []byte("cmd:")) {
		return fmt.Errorf("unrecognized msg: %s", msg)
	} else if bytes.Equal(msg, []byte("cmd:close")) {
		p.shouldClose = true
		return fmt.Errorf("peer requested close")
	}
	command := string(msg)[4:]
	p.conn.WriteString("ack:" + command)
	if p.conn.HasErr() {
		return p.conn.TimeoutErrOrPanic()
	}

	if command == "ping" {
		return nil

	} else {
		return fmt.Errorf("unrecognized command: %s", command)
	}
}
