package peer

import (
	"errors"
	"fmt"

	"github.com/levilutz/basiccoin/src/mainbus"
	"github.com/levilutz/basiccoin/src/util"
)

// Whether we should connect, based on their hello info
func shouldConnect(helloMsg HelloPeerMessage) bool {
	// Don't connect to self
	if helloMsg.RuntimeID == util.Constants.RuntimeID {
		return false
	}
	// Don't connect if version incompatible
	if helloMsg.Version != util.Constants.Version {
		return false
	}
	// TODO: Don't connect if peer already known
	return true
}

// Transmit continue|close, and receive their continue|close. Return whether both peers
// want to continue the connection.
func verifyConnWanted(pc *PeerConn, helloMsg HelloPeerMessage) (bool, error) {
	// Decide if we want to continue and tell them
	shouldConn := shouldConnect(helloMsg)
	if shouldConn {
		pc.TransmitStringLine("continue")
	} else {
		pc.TransmitStringLine("close")
	}

	// Receive whether they want to continue
	contMsg := pc.RetryReadLine(7)
	if err := pc.Err(); err != nil {
		return false, err
	} else if string(contMsg) == "continue" {
		return true, nil
	} else if string(contMsg) == "close" {
		return false, nil
	} else {
		return false, fmt.Errorf("expected 'continue'|'close', received '%s'", contMsg)
	}
}

func GreetPeer(pc *PeerConn, mainBus *mainbus.MainBus) (*PeerBus, error) {
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

	// Close if either peer wants
	conWanted, err := verifyConnWanted(pc, helloMsg)
	if err != nil {
		return nil, err
	}
	if !conWanted {
		return nil, errors.New("peer does not want connection")
	}

	bus := NewPeerBus(helloMsg.RuntimeID, 100)
	go PeerRoutine(pc, bus, mainBus, helloMsg)
	return bus, nil
}

func ReceivePeerGreeting(pc *PeerConn, mainBus *mainbus.MainBus) (*PeerBus, error) {
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

	// Close if either peer wants
	conWanted, err := verifyConnWanted(pc, helloMsg)
	if err != nil {
		return nil, err
	}
	if !conWanted {
		return nil, errors.New("peer does not want connection")
	}

	bus := NewPeerBus(helloMsg.RuntimeID, 100)
	go PeerRoutine(pc, bus, mainBus, helloMsg)
	return bus, nil
}

func PeerRoutine(
	pc *PeerConn, bus *PeerBus, mainBus *mainbus.MainBus, data HelloPeerMessage,
) {
	defer func() {
		// TODO: signal peer dead on bus
		if r := recover(); r != nil {
			fmt.Printf("Failed PeerRoutine: %v\n", r)
		}
	}()
	fmt.Println("Successful connection to:")
	util.PrettyPrint(data)
	// Should be less of a dance here (shouldn't need ConsumeExpected)
	// We emit things, and respond to requests. Is memory/state rly necessary? hope not
	// Loop select new messages in channel, messages from bus channel, ping timer
}
