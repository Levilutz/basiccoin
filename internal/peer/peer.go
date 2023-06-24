package peer

import (
	"bytes"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/levilutz/basiccoin/internal/bus"
	"github.com/levilutz/basiccoin/internal/inv"
	"github.com/levilutz/basiccoin/pkg/core"
	"github.com/levilutz/basiccoin/pkg/prot"
	"github.com/levilutz/basiccoin/pkg/topic"
)

var errPeerClosed = fmt.Errorf("peer requested close")

// The peer's subscriptions.
// Ensure each of these is initialized in NewPeer.
type subscriptions struct {
	PrintUpdate        *topic.SubCh[bus.PrintUpdateEvent]
	SendPeers          *topic.SubCh[bus.SendPeersEvent]
	ShouldAnnounceAddr *topic.SubCh[bus.ShouldAnnounceAddrEvent]
	ShouldRequestPeers *topic.SubCh[bus.ShouldRequestPeersEvent]
	ValidatedHead      *topic.SubCh[bus.ValidatedHeadEvent]
	ValidatedTx        *topic.SubCh[bus.ValidatedTxEvent]
}

// Close our subscriptions as we close.
func (s subscriptions) Close() {
	s.PrintUpdate.Close()
	s.SendPeers.Close()
	s.ShouldAnnounceAddr.Close()
	s.ShouldRequestPeers.Close()
	s.ValidatedHead.Close()
	s.ValidatedTx.Close()
}

// A connection to a single peer.
type Peer struct {
	bus         *bus.Bus
	inv         inv.InvReader
	subs        *subscriptions
	conn        *prot.Conn
	shouldClose bool
	curHead     core.HashT
}

// Create a new peer given a message bus instance.
func NewPeer(msgBus *bus.Bus, inv inv.InvReader, conn *prot.Conn, curHead core.HashT) *Peer {
	subs := &subscriptions{
		PrintUpdate:        msgBus.PrintUpdate.SubCh(),
		SendPeers:          msgBus.SendPeers.SubCh(),
		ShouldAnnounceAddr: msgBus.ShouldAnnounceAddr.SubCh(),
		ShouldRequestPeers: msgBus.ShouldRequestPeers.SubCh(),
		ValidatedHead:      msgBus.ValidatedHead.SubCh(),
		ValidatedTx:        msgBus.ValidatedTx.SubCh(),
	}
	return &Peer{
		bus:         msgBus,
		inv:         inv,
		subs:        subs,
		conn:        conn,
		shouldClose: false,
		curHead:     curHead,
	}
}

// Start the peer's loop.
func (p *Peer) Loop() {
	// Handle panics and unsubscribe
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("peer %s closed from panic: %s\n", p.conn.PeerRuntimeId(), r)
			debug.PrintStack()
		} else {
			fmt.Printf("peer %s closed\n", p.conn.PeerRuntimeId())
		}
		p.subs.Close()
		p.bus.PeerClosing.Pub(bus.PeerClosingEvent{
			PeerRuntimeId: p.conn.PeerRuntimeId(),
		})
	}()

	if p.conn.WeAreInitiator() {
		p.issueCommandPrintErr(syncChainCmd, p.handleSyncChain)
	}

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
				return p.handleWritePeerAddrs(event.PeerAddrs)
			})

		case event := <-p.subs.ShouldAnnounceAddr.C:
			if event.TargetRuntimeId != p.conn.PeerRuntimeId() {
				continue
			}
			p.issueCommandPrintErr(announceAddrCmd, func() error {
				return p.handleWriteAnnounceAddr(event.Addr)
			})

		case event := <-p.subs.ValidatedTx.C:
			p.issueCommandPrintErr(newTxCmd, func() error {
				return p.handleWriteNewTx(event.TxId)
			})

		case event := <-p.subs.ValidatedHead.C:
			p.curHead = event.Head
			p.issueCommandPrintErr(syncChainCmd, p.handleSyncChain)

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

	} else if command == newTxCmd {
		return p.handleReadNewTx()

	} else if command == syncChainCmd {
		return p.handleSyncChain()

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
