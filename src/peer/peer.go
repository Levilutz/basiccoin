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

func (p *Peer) SetHead(head db.HashT) {
	go func() {
		p.EventBus <- events.NewHeadPeerEvent{
			Head: head,
		}
	}()
}

// Loop handling events from our message bus and the peer.
func (p *Peer) Loop() {
	defer fmt.Println("Peer closed:", p.HelloMsg.RuntimeID)
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
			if err := p.conn.Err(); err != nil {
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
		return true, p.handleClose(true, false)

	case events.NewHeadPeerEvent:
		p.head = msg.Head
		// TODO: Inform the peer of our head block

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

	case events.BroadcastBlockPeerEvent:
		return false, handleTransmitNewBlockExchange(msg.BlockId, p.conn, p.inv)

	default:
		fmt.Printf("Unhandled peer event %T\n", event)
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
		return true, p.handleClose(false, true)
	}

	p.conn.TransmitStringLine("ack:" + command)
	if err := p.conn.Err(); err != nil {
		return false, err
	}

	if command == "ping" {

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

	} else if command == "new-block" {
		mainBusEvent, err := handleReceiveNewBlockExchange(p.conn, p.inv)
		if err != nil {
			return false, err
		}
		go func() {
			p.mainBus <- mainBusEvent
		}()

	} else {
		fmt.Println("Unexpected peer message:", command)
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
	if err := p.conn.Err(); err != nil {
		return false, err
	}
	// Happy path - they acknowledged us
	if string(resp) == "ack:"+command {
		return false, handler()
	}
	// Sad path - we sent commands simultaneously
	if bytes.HasPrefix(resp, []byte("cmd:")) {
		if string(resp) == "cmd:close" {
			// If their command was a close, handle it immediately
			return true, p.handleClose(false, true)

		} else if p.weAreInitiator {
			// If we initiated the og handshake, honor their cmd, then expect ours to be
			shouldClose, err := p.handleReceivedLine(resp)
			if shouldClose || err != nil {
				return shouldClose, err
			}
			p.conn.ConsumeExpected("ack:" + command)
			if err := p.conn.Err(); err != nil {
				return false, err
			}
			return false, handler()

		} else {
			// If we received the og handshake, expect to be honored, then honor theirs
			p.conn.ConsumeExpected("ack:" + command)
			if err := p.conn.Err(); err != nil {
				return false, err
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

func (p *Peer) handleClose(issuing bool, notifyMainBus bool) error {
	if issuing {
		p.conn.TransmitStringLine("cmd:close")
	}
	if notifyMainBus {
		go func() {
			p.mainBus <- events.PeerClosingMainEvent{
				RuntimeID: p.HelloMsg.RuntimeID,
			}
		}()
	}
	if err := p.conn.Err(); err != nil {
		return err
	}
	return p.conn.Close()
}
