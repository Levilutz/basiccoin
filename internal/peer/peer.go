package peer

import (
	"bytes"
	"fmt"
	"time"

	"github.com/levilutz/basiccoin/internal/inv"
	"github.com/levilutz/basiccoin/internal/pubsub"
	"github.com/levilutz/basiccoin/pkg/prot"
	"github.com/levilutz/basiccoin/pkg/topic"
)

var errPeerClosed = fmt.Errorf("peer requested close")

// The peer's subscriptions.
// Ensure each of these is initialized in NewPeer.
type subscriptions struct {
	PrintUpdate        *topic.SubCh[pubsub.PrintUpdateEvent]
	SendPeers          *topic.SubCh[pubsub.SendPeersEvent]
	ShouldAnnounceAddr *topic.SubCh[pubsub.ShouldAnnounceAddrEvent]
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
	inv         inv.InvReader
	subs        *subscriptions
	conn        *prot.Conn
	shouldClose bool
}

// Create a new peer given a message bus instance.
func NewPeer(pubSub *pubsub.PubSub, inv inv.InvReader, conn *prot.Conn) *Peer {
	subs := &subscriptions{
		PrintUpdate:        pubSub.PrintUpdate.SubCh(),
		SendPeers:          pubSub.SendPeers.SubCh(),
		ShouldAnnounceAddr: pubSub.ShouldAnnounceAddr.SubCh(),
		ShouldRequestPeers: pubSub.ShouldRequestPeers.SubCh(),
		ValidatedHead:      pubSub.ValidatedHead.SubCh(),
	}
	return &Peer{
		pubSub: pubSub,
		inv:    inv,
		subs:   subs,
		conn:   conn,
	}
}

// Start the peer's loop.
func (p *Peer) Loop() {
	// Handle panics and unsubscribe
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("peer %s closed from panic: %s\n", p.conn.PeerRuntimeId(), r)
		} else {
			fmt.Printf("peer %s closed\n", p.conn.PeerRuntimeId())
		}
		p.pubSub.PeerClosing.Pub(pubsub.PeerClosingEvent{
			PeerRuntimeId: p.conn.PeerRuntimeId(),
		})
	}()

	// Loop
	for {
		if p.shouldClose {
			return
		}
		select {
		case event := <-p.subs.ShouldRequestPeers.C:
			if event.TargetRuntimeId != p.conn.PeerRuntimeId() {
				continue
			}
			p.issueCommandPrintErr(addrsRequestCmd, p.handleWriteAddrsRequest)

		case event := <-p.subs.SendPeers.C:
			if event.TargetRuntimeId != p.conn.PeerRuntimeId() {
				continue
			}
			p.issueCommandPrintErr(peerAddrsCmd, func() error {
				return p.handleWritePeerAddrs(event)
			})

		case event := <-p.subs.ShouldAnnounceAddr.C:
			if event.TargetRuntimeId != p.conn.PeerRuntimeId() {
				continue
			}
			p.issueCommandPrintErr(announceAddrCmd, func() error {
				return p.handleWriteAnnounceAddr(event)
			})

		case event := <-p.subs.ValidatedHead.C:
			fmt.Println("new validated head:", event.Head)

		case event := <-p.subs.PrintUpdate.C:
			if !event.Peer {
				continue
			}
			fmt.Printf("peer exists: %s\n", p.conn.PeerRuntimeId())

		default:
			msg := p.conn.ReadTimeout(time.Millisecond * 100)
			if p.conn.TimeoutErrOrPanic() != nil {
				continue
			}
			if err := p.handleReceivedMessage(msg); err != nil {
				fmt.Printf("error handling '%s': %s\n", msg, err.Error())
			}
		}
	}
}

// Handle a message received from a peer.
func (p *Peer) handleReceivedMessage(msg []byte) error {
	if !bytes.HasPrefix(msg, []byte("cmd:")) {
		return fmt.Errorf("unrecognized msg: %s", msg)
	} else if bytes.Equal(msg, []byte("cmd:close")) {
		p.shouldClose = true
		return errPeerClosed
	}
	command := string(msg)[4:]
	p.conn.WriteString("ack:" + command)
	if p.conn.HasErr() {
		return p.conn.Err()
	}

	if command == "ping" {
		return nil

	} else if command == addrsRequestCmd {
		return p.handleReadAddrsRequest()

	} else if command == peerAddrsCmd {
		return p.handleReadPeerAddrs()

	} else if command == announceAddrCmd {
		return p.handleReadAnnounceAddr()

	} else {
		return fmt.Errorf("unrecognized command: %s", command)
	}
}

// Issue an outbound command with the given handler, print err instead of returning.
func (p *Peer) issueCommandPrintErr(command string, handler func() error) {
	err := p.issueCommand(command, handler)
	if err != nil {
		fmt.Printf("error issuing %s: %s\n", command, err.Error())
	}
}

// Issue an outbound command with the given handler.
func (p *Peer) issueCommand(command string, handler func() error) error {
	p.conn.WriteString("cmd:" + command)
	// Expect to receive either 'ack:ourCommand' or 'cmd:theirCommand'
	resp := p.conn.Read()
	if p.conn.HasErr() {
		return p.conn.Err()
	} else if bytes.Equal(resp, []byte("cmd:close")) {
		p.shouldClose = true
		return errPeerClosed
	}
	// Happy path - they acknowledged us
	if string(resp) == "ack:"+command {
		return handler()
	}
	// Other path - we sent commands simultaneously
	if bytes.HasPrefix(resp, []byte("cmd:")) {
		if p.conn.WeAreInitiator() {
			// If we initiated the original connection, honor their command first
			if err := p.handleReceivedMessage(resp); err != nil {
				return err
			} else if p.shouldClose {
				return errPeerClosed
			}
			p.conn.ReadStringExpected("ack:" + command)
			if p.conn.HasErr() {
				return p.conn.Err()
			}
			return handler()
		} else {
			// If they initiated the original connnection, honor our command first
			p.conn.ReadStringExpected("ack:" + command)
			if p.conn.HasErr() {
				return p.conn.Err()
			}
			if err := handler(); err != nil {
				return err
			} else if p.shouldClose {
				return errPeerClosed
			}
			return p.handleReceivedMessage(resp)
		}
	}
	// Neither
	return fmt.Errorf("unrecognized msg: %s", resp)
}
