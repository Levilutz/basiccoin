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
	HelloMsg       *HelloPeerMessage
	EventBus       chan events.PeerEvent
	conn           *PeerConn
	mainBus        chan<- events.MainEvent
	weAreInitiator bool
}

// Create a Peer from the result of a successfull handshake on a PeerConn, the
// associated PeerConn, and a bus to emit events back to the manager loop.
func NewPeer(
	msg *HelloPeerMessage,
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

	return NewPeer(helloMsg, pc, mainBus, true), nil
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
	var err error
	listenTicker := time.NewTicker(util.Constants.PeerListenFreq)
	pingTicker := time.NewTicker(util.Constants.PeerPingFreq)
	for {
		shouldClose := false
		select {
		case event := <-p.EventBus:
			shouldClose = p.handlePeerBusEvent(event)
		case <-listenTicker.C:
			line := p.conn.ReadLineTimeout(25)
			if p.conn.Err() != nil {
				continue
			}
			shouldClose, err = p.handleReceivedLine(line)
			if err != nil {
				fmt.Printf("error handling '%s': %s\n", line, err.Error())
			}
		case <-pingTicker.C:
			shouldClose, err = p.issuePeerCommand("ping", func() error {
				return nil
			})
			if err != nil {
				fmt.Println("peer lost:", p.HelloMsg.RuntimeID, err.Error())
				return
			}
		}
		if shouldClose {
			return
		}
	}
}

// Handle event from our message bus, return whether we should close.
func (p *Peer) handlePeerBusEvent(event events.PeerEvent) bool {
	fmt.Println(event)
	return false
}

// Handle command received from peer, returns whether we should close.
func (p *Peer) handleReceivedLine(line []byte) (bool, error) {
	if !bytes.HasPrefix(line, []byte("cmd:")) {
		return false, fmt.Errorf("unrecognized line: %s", line)
	}
	command := string(line)[4:]
	// TODO ack before handling
	if command == "close" {
		return true, p.handleClose()

	} else if command == "ping" {
		p.conn.TransmitStringLine("ack:ping")
		return false, p.conn.Err()

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
			return true, p.handleClose()

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

func (p *Peer) handleClose() error {
	p.conn.TransmitStringLine("ack:close")
	p.mainBus <- events.MainEvent{
		PeerClosing: &events.PeerClosingMainEvent{
			RuntimeID: p.HelloMsg.RuntimeID,
		},
	}
	return p.conn.Err()
}
