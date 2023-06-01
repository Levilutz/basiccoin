package peer

import (
	"bytes"
	"fmt"
	"time"

	"github.com/levilutz/basiccoin/src/db"
	"github.com/levilutz/basiccoin/src/events"
	"github.com/levilutz/basiccoin/src/util"
)

// Encapsulate a high-level connection to a peer.
type Peer struct {
	HelloMsg       *HelloMessage
	EventBus       chan any // TODO: Make event bus private
	conn           *PeerConn
	mainBus        chan<- any
	weAreInitiator bool
	inv            db.InvReader
	head           db.HashT
}

// Create a Peer.
// "msg" is the result of a successful handshake on a PeerConn.
// "pc" is the peerconn on which we have already handshaked.
// "mainBus" is a bus on which to emit events back to the manager.
// "weAreInitiator" is whether we are the peer that initiated the connection.
// "inv" is a InvReader.
func NewPeer(
	msg *HelloMessage,
	pc *PeerConn,
	mainBus chan any,
	weAreInitiator bool,
	inv db.InvReader,
	head db.HashT,
) *Peer {
	return &Peer{
		HelloMsg:       msg,
		EventBus:       make(chan any),
		conn:           pc,
		mainBus:        mainBus,
		weAreInitiator: weAreInitiator,
		inv:            inv,
		head:           head,
	}
}

func (p *Peer) SyncHead(head db.HashT) {
	go func() {
		p.EventBus <- events.SyncHeadPeerEvent{
			Head: head,
		}
	}()
}

// Loop handling events from our message bus and the peer.
func (p *Peer) Loop() {
	defer func() {
		go func() {
			fmt.Println("peer closed:", p.HelloMsg.RuntimeID)
			p.mainBus <- events.PeerClosingMainEvent{
				RuntimeID: p.HelloMsg.RuntimeID,
			}
		}()
	}()
	var err error
	pingTicker := time.NewTicker(util.Constants.PeerPingFreq)
	for {
		shouldClose := false
		select {
		case event := <-p.EventBus:
			shouldClose, err = p.handlePeerBusEvent(event)
			if err != nil {
				fmt.Printf("error handling event '%v': %s\n", event, err.Error())
			}

		case <-pingTicker.C:
			shouldClose, err = p.issuePeerCommand("ping", func() error {
				return nil
			})
			if err != nil {
				fmt.Println("peer lost:", p.HelloMsg.RuntimeID, err.Error())
				return
			}

		default:
			line := p.conn.ReadLineTimeout(100 * time.Millisecond)
			if p.conn.HasErr() {
				continue
			}
			shouldClose, err = p.handleReceivedLine(line)
			if err != nil {
				fmt.Printf("error handling line '%s': %s\n", line, err.Error())
			}
		}
		if shouldClose {
			return
		}
	}
}

// Handle event from our message bus, return whether we should close.
func (p *Peer) handlePeerBusEvent(event any) (bool, error) {
	switch msg := event.(type) {
	case events.ShouldEndPeerEvent:
		return true, p.handleClose(true)

	case events.SyncHeadPeerEvent:
		p.head = msg.Head
		return p.issuePeerCommand("sync", p.handleSync)

	case events.PeersDataPeerEvent:
		return p.issuePeerCommand("addrs", func() error {
			p.conn.TransmitMessage(AddrsMessage{
				PeerAddrs: msg.PeerAddrs,
			})
			return p.conn.Err()
		})

	case events.PeersWantedPeerEvent:
		return p.issuePeerCommand("peers-wanted", func() error {
			return nil
		})

	default:
		fmt.Printf("unhandled peer event %T\n", event)
	}
	return false, nil
}

// Handle command received from peer, returns whether we should close.
func (p *Peer) handleReceivedLine(line []byte) (bool, error) {
	if !bytes.HasPrefix(line, []byte("cmd:")) {
		return false, fmt.Errorf("unrecognized line: %s", line)
	}
	command := string(line)[4:]
	if command == "close" {
		return true, p.handleClose(false)
	}

	p.conn.TransmitStringLine("ack:" + command)
	if p.conn.HasErr() {
		return false, p.conn.Err()
	}

	if command == "ping" {
		// ack above was sufficient

	} else if command == "sync" {
		// handleSync sends to main bus if appropriate
		if err := p.handleSync(); err != nil {
			return false, err
		}

	} else if command == "addrs" {
		msg, err := ReceiveAddrsMessage(p.conn)
		if err != nil {
			return false, err
		}
		go func() {
			p.mainBus <- events.PeersReceivedMainEvent{
				PeerAddrs: msg.PeerAddrs,
			}
		}()

	} else if command == "peers-wanted" {
		go func() {
			p.mainBus <- events.PeersWantedMainEvent{
				PeerRuntimeID: p.HelloMsg.RuntimeID,
			}
		}()

	} else {
		fmt.Println("unexpected peer message:", command)
	}

	return false, nil
}

// Issue an outbound interaction for the command (given without "cmd:").
// Handler is what to run after they ack. Returns whether we should close.
// If us and peer simultaneously issued commands, the og handshake initiator goes last.
func (p *Peer) issuePeerCommand(command string, handler func() error) (bool, error) {
	p.conn.TransmitStringLine("cmd:" + command)
	// Expect to receive either "ack:our command" or "cmd:their command"
	resp := p.conn.RetryReadLine(7)
	if p.conn.HasErr() {
		return false, p.conn.Err()
	}
	// Happy path - they acknowledged us
	if string(resp) == "ack:"+command {
		return false, handler()
	}
	// Sad path - we sent commands simultaneously
	if bytes.HasPrefix(resp, []byte("cmd:")) {
		if string(resp) == "cmd:close" {
			// If their command was a close, handle it immediately
			return true, p.handleClose(false)

		} else if p.weAreInitiator {
			// If we initiated the og handshake, honor their cmd, then expect ours to be
			shouldClose, err := p.handleReceivedLine(resp)
			if shouldClose || err != nil {
				return shouldClose, err
			}
			p.conn.ConsumeExpected("ack:" + command)
			if p.conn.HasErr() {
				return false, p.conn.Err()
			}
			return false, handler()

		} else {
			// If we received the og handshake, expect to be honored, then honor theirs
			p.conn.ConsumeExpected("ack:" + command)
			if p.conn.HasErr() {
				return false, p.conn.Err()
			}
			err := handler()
			if err != nil {
				return false, err
			}
			return p.handleReceivedLine(resp)
		}
	}
	return false, nil
}

func (p *Peer) handleClose(issuing bool) error {
	if issuing {
		p.conn.TransmitStringLine("cmd:close")
	}
	if p.conn.HasErr() {
		return p.conn.Err()
	}
	return p.conn.Close()
}

// Handle a sync, inbound or outbound.
// If we're successful and we received the sync, send InboundSyncMainEvent to main.
// If we're successful and we sent the sync or no sync occurred, send nothing to main.
func (p *Peer) handleSync() error {
	ourWork := p.inv.GetBlockTotalWork(p.head)
	p.conn.TransmitHashLine(ourWork)
	theirWork := p.conn.RetryReadHashLine(7)
	if p.conn.HasErr() {
		return p.conn.Err()
	}
	if theirWork == ourWork {
		// Neither peer wants to sync
		p.conn.TransmitStringLine("fin:sync")
		p.conn.ConsumeExpected("fin:sync")
		return p.conn.Err()
	} else if db.HashLT(theirWork, ourWork) {
		// Send a sync
		p.conn.TransmitStringLine("sync-send")
		p.conn.ConsumeExpected("sync-recv")
		if p.conn.HasErr() {
			return p.conn.Err()
		}
		return p.handleSendSync()
	} else {
		// Receive a sync
		p.conn.TransmitStringLine("sync-recv")
		p.conn.ConsumeExpected("sync-send")
		if p.conn.HasErr() {
			return p.conn.Err()
		}
		eventP, err := p.handleReceiveSync()
		if err != nil {
			return err
		}
		go func() {
			p.mainBus <- eventP
		}()
		return nil
	}
}

// Send a chain sync.
func (p *Peer) handleSendSync() error {
	// Find last common ancestor with peer
	neededBlockIds := make([]db.HashT, 0)
	lcaId := p.head
	p.conn.TransmitHashLine(lcaId)
	resp := p.conn.RetryReadStringLine(7)
	if p.conn.HasErr() {
		return p.conn.Err()
	}
	for resp == "next" {
		neededBlockIds = append(neededBlockIds, lcaId)
		lcaId = p.inv.GetBlockParentId(lcaId)
		p.conn.TransmitHashLine(lcaId)
		resp = p.conn.RetryReadStringLine(7)
		if p.conn.HasErr() {
			return p.conn.Err()
		}
	}
	if resp != "recognized" {
		return fmt.Errorf("expected 'recognized', received %s", resp)
	}
	if len(neededBlockIds) == 0 {
		panic("ruh roh")
	}
	// Send blocks to peer
	// Check if peer verified the chain's work
	// Until peer says "inv-complete", receive what entities they want, and send them
	return nil
}

// Receive a chain sync.
func (p *Peer) handleReceiveSync() (*events.InboundSyncMainEvent, error) {
	// Find last common ancestor with peer
	newHead := p.conn.RetryReadHashLine(7)
	lcaId := newHead
	for !p.inv.HasBlock(lcaId) {
		p.conn.TransmitStringLine("next")
		lcaId = p.conn.RetryReadHashLine(7)
	}
	p.conn.TransmitStringLine("recognized")
	// Receive blocks from peer
	// Verify the chain's continuity, attaching to ours, and work
	// Until our inv / local inv is complete, request entities, and add to table
	// Build the event to send to manager
	return &events.InboundSyncMainEvent{
		Head:    db.HashTZero,
		Blocks:  nil,
		Merkles: nil,
		Txs:     nil,
	}, nil
}
