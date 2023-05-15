package peer

import (
	"bytes"
	"fmt"
	"net"
	"time"

	"github.com/levilutz/basiccoin/src/events"
	"github.com/levilutz/basiccoin/src/util"
)

// Encapsulate a high-level connection to a peer.
type Peer struct {
	HelloMsg       *HelloMessage
	EventBus       chan events.PeerEvent
	conn           *PeerConn
	mainBus        chan<- events.MainEvent
	weAreInitiator bool
}

// Create a Peer from the result of a successfull handshake on a PeerConn, the
// associated PeerConn, and a bus to emit events back to the manager loop.
func NewPeer(
	msg *HelloMessage,
	pc *PeerConn,
	mainBus chan events.MainEvent,
	weAreInitiator bool,
) *Peer {
	return &Peer{
		HelloMsg:       msg,
		EventBus:       make(chan events.PeerEvent, util.Constants.PeerBusBufferSize),
		conn:           pc,
		mainBus:        mainBus,
		weAreInitiator: weAreInitiator,
	}
}

// Attempt to initialize an outbound connection given a remote address.
func NewPeerOutbound(addr string, mainBus chan events.MainEvent) (*Peer, error) {
	// Resolve host
	pc, err := ResolvePeerConn(addr)
	if err != nil {
		return nil, err
	}

	// Hello handshake
	helloMsg := pc.GiveHandshake()
	if err := pc.Err(); err != nil {
		return nil, err
	}

	peer := NewPeer(helloMsg, pc, mainBus, true)

	// Advertise ourselves if wanted
	if util.Constants.Listen {
		if _, err := peer.issuePeerCommand("addrs", func() error {
			peer.conn.TransmitMessage(AddrsMessage{
				PeerAddrs: []string{util.Constants.LocalAddr},
			})
			return peer.conn.Err()
		}); err != nil {
			return nil, err
		}
	}

	return peer, nil
}

// Attempt to initialize a new inbound connection given the TCP Conn.
func NewPeerInbound(conn *net.TCPConn, mainBus chan events.MainEvent) (*Peer, error) {
	// Make PeerConn
	pc := NewPeerConn(conn)

	// Hello handshake
	helloMsg := pc.ReceiveHandshake()
	if err := pc.Err(); err != nil {
		return nil, err
	}

	return NewPeer(helloMsg, pc, mainBus, false), nil
}

// Loop handling events from our message bus and the peer
func (p *Peer) Loop() {
	if util.Constants.DebugPeerLoop {
		fmt.Println("PEER_LOOP")
	}
	defer fmt.Println("Peer closed:", p.HelloMsg.RuntimeID)
	var err error
	pingTicker := time.NewTicker(util.Constants.PeerPingFreq)
	for {
		shouldClose := false
		select {
		case event := <-p.EventBus:
			if util.Constants.DebugPeerLoop {
				fmt.Println("PEER_EVENT", event)
			}
			shouldClose, err = p.handlePeerBusEvent(event)
			if err != nil {
				fmt.Printf("error handling event '%v': %s\n", event, err.Error())
			}

		case <-pingTicker.C:
			if util.Constants.DebugPeerLoop {
				fmt.Println("PEER_PING")
			}
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
			if util.Constants.DebugPeerLoop {
				fmt.Println("PEER_LISTEN", string(line))
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
func (p *Peer) handlePeerBusEvent(event events.PeerEvent) (bool, error) {
	if msg := event.ShouldEnd; msg != nil {
		return true, p.handleClose(true, false)

	} else if msg := event.PeersData; msg != nil {
		return p.issuePeerCommand("addrs", func() error {
			p.conn.TransmitMessage(AddrsMessage{
				PeerAddrs: msg.PeerAddrs,
			})
			return p.conn.Err()
		})
	}
	return false, nil
}

// Handle command received from peer, returns whether we should close.
func (p *Peer) handleReceivedLine(line []byte) (bool, error) {
	if !bytes.HasPrefix(line, []byte("cmd:")) {
		return false, fmt.Errorf("unrecognized line: %s", line)
	}
	command := string(line)[4:]
	// TODO ack before handling
	if command == "close" {
		return true, p.handleClose(false, true)

	} else if command == "ping" {
		p.conn.TransmitStringLine("ack:ping")
		return false, p.conn.Err()

	} else if command == "addrs" {
		p.conn.TransmitStringLine("ack:addrs")
		msg, err := ReceiveAddrsMessage(p.conn)
		if err != nil {
			return false, err
		}
		p.mainBus <- events.MainEvent{
			PeersReceived: &events.PeersReceivedMainEvent{
				PeerAddrs: msg.PeerAddrs,
			},
		}
		return false, nil

	} else {
		fmt.Println("Unexpected peer message:", command)
		return false, nil
	}
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
			p.handleReceivedLine(resp)
		}
	}
	return false, nil
}

func (p *Peer) handleClose(issuing bool, notifyMainBus bool) error {
	if issuing {
		p.conn.TransmitStringLine("cmd:close")
	}
	if notifyMainBus {
		p.mainBus <- events.MainEvent{
			PeerClosing: &events.PeerClosingMainEvent{
				RuntimeID: p.HelloMsg.RuntimeID,
			},
		}
	}
	return p.conn.Err()
}
