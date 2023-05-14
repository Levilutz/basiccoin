package peer

import (
	"errors"
	"fmt"
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
	msg *HelloPeerMessage, pc *PeerConn, bufferSize int, mainBus *mainbus.MainBus,
) *Peer {
	return &Peer{
		HelloMsg: msg,
		EventBus: make(chan PeerEvent, bufferSize),
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
	pc.TransmitMessage(NewHelloMessage())
	pc.ConsumeExpected("ack:hello")
	pc.ConsumeExpected("hello")
	if err := pc.Err(); err != nil {
		return nil, err
	}
	helloMsg, err := ReceiveHelloMessage(pc)
	if err != nil {
		return nil, err
	}
	pc.TransmitStringLine("ack:hello")
	if err = pc.Err(); err != nil {
		return nil, err
	}

	p := NewPeer(&helloMsg, pc, 100, mainBus)

	// Close if either peer wants
	conWanted, err := p.verifyConnWanted()
	if err != nil {
		return nil, err
	}
	if !conWanted {
		return nil, errors.New("peer does not want connection")
	}

	return p, nil
}

// Whether we should connect, based on their hello info
func (p *Peer) shouldConnect() bool {
	// Don't connect to self
	if p.HelloMsg.RuntimeID == util.Constants.RuntimeID {
		return false
	}
	// Don't connect if version incompatible
	if p.HelloMsg.Version != util.Constants.Version {
		return false
	}
	// TODO: Don't connect if peer already known
	return true
}

// Transmit continue|close, and receive their continue|close. Return whether both peers
// want to continue the connection.
func (p *Peer) verifyConnWanted() (bool, error) {
	// Decide if we want to continue and tell them
	if p.shouldConnect() {
		p.conn.TransmitStringLine("continue")
	} else {
		p.conn.TransmitStringLine("close")
	}

	// Receive whether they want to continue
	contMsg := p.conn.RetryReadLine(7)
	if err := p.conn.Err(); err != nil {
		return false, err
	} else if string(contMsg) == "continue" {
		return true, nil
	} else if string(contMsg) == "close" {
		return false, nil
	} else {
		return false, fmt.Errorf("expected 'continue'|'close', received '%s'", contMsg)
	}
}

func ReceivePeerGreeting(pc *PeerConn, mainBus *mainbus.MainBus) (*Peer, error) {
	// Hello handshake
	pc.ConsumeExpected("hello")
	if err := pc.Err(); err != nil {
		return nil, err
	}
	helloMsg, err := ReceiveHelloMessage(pc)
	if err != nil {
		return nil, err
	}
	pc.TransmitStringLine("ack:hello")
	pc.TransmitMessage(NewHelloMessage())
	pc.ConsumeExpected("ack:hello")
	if err := pc.Err(); err != nil {
		return nil, err
	}

	p := NewPeer(&helloMsg, pc, 100, mainBus)

	// Close if either peer wants
	conWanted, err := p.verifyConnWanted()
	if err != nil {
		return nil, err
	}
	if !conWanted {
		return nil, errors.New("peer does not want connection")
	}

	go p.Loop()
	return p, nil
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
			line, err := p.conn.ReadLineTimeout(25)
			if err != nil {
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
