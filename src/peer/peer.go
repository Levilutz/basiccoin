package peer

import (
	"fmt"
	"net"
	"time"

	"github.com/levilutz/basiccoin/src/mainbus"
	"github.com/levilutz/basiccoin/src/util"
)

type Peer struct {
	HelloMsg *HelloPeerMessage
	EventBus chan PeerEvent
	conn     *PeerConn
	mainBus  *mainbus.MainBus
}

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

func (p *Peer) Loop() {
	defer func() {
		// TODO: signal peer dead on bus
		if r := recover(); r != nil {
			fmt.Printf("Failed PeerRoutine: %v\n", r)
		}
	}()
	fmt.Println("Successful connection to:")
	util.PrettyPrint(p.HelloMsg)
	ticker := time.NewTicker(time.Millisecond * time.Duration(100))
	for {
		select {
		case event := <-p.EventBus:
			fmt.Println(event)
		case <-ticker.C:
			line := p.conn.ReadLineTimeout(25)
			if p.conn.Err() != nil {
				continue
			}
			cmd := string(line)

			if cmd == "close" {
				p.conn.TransmitStringLine("close")
				p.mainBus.Events <- mainbus.MainBusEvent{
					PeerClosing: &mainbus.PeerClosingEvent{
						RuntimeID: p.HelloMsg.RuntimeID,
					},
				}
				return

			} else if cmd == "ping" {
				p.conn.TransmitStringLine("pong")

			} else {
				fmt.Println("Unexpected peer message:", cmd)
			}
		}
	}
	// Should be less of a dance here (shouldn't need ConsumeExpected)
	// We emit things, and respond to requests. Is memory/state rly necessary? hope not
	// Loop select new messages in channel, messages from bus channel, ping timer
}
