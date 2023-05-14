package peer

import (
	"fmt"
	"net"
	"time"

	"github.com/levilutz/basiccoin/src/mainbus"
	"github.com/levilutz/basiccoin/src/util"
)

// Encapsulate a high-level connection to a peer.
type Peer struct {
	HelloMsg *HelloPeerMessage
	EventBus chan PeerEvent
	conn     *PeerConn
	mainBus  *mainbus.MainBus
}

// Create a Peer from the result of a successfull handshake on a PeerConn, the
// associated PeerConn, and a bus to emit events back to the manager loop.
func NewPeer(
	msg *HelloPeerMessage, pc *PeerConn, mainBus *mainbus.MainBus,
) *Peer {
	return &Peer{
		HelloMsg: msg,
		EventBus: make(chan PeerEvent, util.Constants.PeerBusBufferSize),
		conn:     pc,
		mainBus:  mainBus,
	}
}

// Attempt to initialize an outbound connection given a remote address.
func NewPeerOutbound(addr string, mainBus *mainbus.MainBus) (*Peer, error) {
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

	return NewPeer(helloMsg, pc, mainBus), nil
}

// Attempt to initialize a new inbound connection given the TCP Conn.
func NewPeerInbound(conn *net.TCPConn, mainBus *mainbus.MainBus) (*Peer, error) {
	// Make PeerConn
	pc := NewPeerConn(conn)

	// Hello handshake
	helloMsg := pc.ReceiveHandshake()
	if err := pc.Err(); err != nil {
		return nil, err
	}

	return NewPeer(helloMsg, pc, mainBus), nil
}

// Loop handling events from our message bus and the peer
func (p *Peer) Loop() {
	listenTicker := time.NewTicker(util.Constants.PeerListenFreq)
	pingTicker := time.NewTicker(util.Constants.PeerPingFreq)
	for {
		select {
		case event := <-p.EventBus:
			p.handlePeerBusEvent(event)
		case <-listenTicker.C:
			line := p.conn.ReadLineTimeout(25)
			if p.conn.Err() != nil {
				continue
			}
			p.handleReceivedCommand(string(line))
		case <-pingTicker.C:
			p.conn.TransmitStringLine("ping")
			p.conn.ConsumeExpected("pong")
			if err := p.conn.Err(); err == nil {
				fmt.Println("ping success")
			} else {
				fmt.Println("ping err", err.Error())
			}
		}
	}
}

// Handle event from our message bus
func (p *Peer) handlePeerBusEvent(event PeerEvent) {
	fmt.Println(event)
}

// Handle command received from peer
func (p *Peer) handleReceivedCommand(command string) {
	if command == "close" {
		p.conn.TransmitStringLine("close")
		p.mainBus.Events <- mainbus.MainBusEvent{
			PeerClosing: &mainbus.PeerClosingEvent{
				RuntimeID: p.HelloMsg.RuntimeID,
			},
		}
		return

	} else if command == "ping" {
		p.conn.TransmitStringLine("pong")

	} else {
		fmt.Println("Unexpected peer message:", command)
	}
}
