package main

import (
	"errors"
	"fmt"

	"github.com/levilutz/basiccoin/src/util"
)

// Whether we should connect, based on their hello info
func shouldConnect(helloMsg HelloMessage) bool {
	// Don't connect to self
	if helloMsg.RuntimeID == util.Constants.RuntimeID {
		return false
	}
	// Don't connect if version incompatible
	if helloMsg.Version != util.Constants.Version {
		// TODO: Actually check compat, not equality
		return false
	}
	// TODO: Don't connect if peer already known
	return true
}

// Transmit continue|close, and receive their continue|close. Return whether both peers
// want to continue the connection.
func verifyConnWanted(pc util.PeerConn, helloMsg HelloMessage) (bool, error) {
	// Decide if we want to continue and tell them
	var err error
	shouldConn := shouldConnect(helloMsg)
	if shouldConn {
		err = TransmitSimpleMessage(pc, "continue")
	} else {
		err = TransmitSimpleMessage(pc, "close")
	}
	if err != nil {
		return false, err
	}

	// Receive whether they want to continue
	contMsg, err := util.RetryReadLine(pc, 8)
	if err != nil {
		return false, err
	}
	theyWantConn := false
	switch string(contMsg) {
	case "continue":
		theyWantConn = true
	case "close":
		theyWantConn = false
	default:
		return false, fmt.Errorf("expected 'continue'|'close', received '%s'", contMsg)
	}

	return theyWantConn, nil
}

func GreetPeer(pc util.PeerConn) error {
	defer func() {
		if r := recover(); r != nil {
			// TODO: signal peer dead on bus
			fmt.Printf("Failed GreetPeer: %v\n", r)
		}
	}()

	// Hello handshake
	err := NewHelloMessage().Transmit(pc)
	if err != nil {
		// TODO: signal peer dead on bus
		return err
	}
	err = ConsumeExpected(pc, "ack:hello")
	if err != nil {
		// TODO: signal peer dead on bus
		return err
	}
	err = ConsumeExpected(pc, "hello")
	if err != nil {
		// TODO: signal peer dead on bus
		return err
	}
	helloMsg, err := ReceiveHelloMessage(pc)
	if err != nil {
		// TODO: signal peer dead on bus
		return err
	}
	err = TransmitSimpleMessage(pc, "ack:hello")
	if err != nil {
		// TODO: signal peer dead on bus
		return err
	}

	// Close if either peer wants
	conWanted, err := verifyConnWanted(pc, helloMsg)
	if err != nil {
		// TODO: signal peer dead on bus
		return err
	}
	if !conWanted {
		// TODO: signal peer dead on bus
		return errors.New("peer does not want connection")
	}

	go PeerRoutine(pc, helloMsg)
	return nil
}

func ReceivePeerGreeting(pc util.PeerConn) {
	defer func() {
		if r := recover(); r != nil {
			// TODO: signal peer dead on bus
			fmt.Printf("Failed ReceivePeerGreeting: %v\n", r)
		}
	}()

	// Hello handshake
	err := ConsumeExpected(pc, "hello")
	util.PanicErr(err)
	helloMsg, err := ReceiveHelloMessage(pc)
	util.PanicErr(err)
	err = TransmitSimpleMessage(pc, "ack:hello")
	util.PanicErr(err)
	err = NewHelloMessage().Transmit(pc)
	util.PanicErr(err)
	err = ConsumeExpected(pc, "ack:hello")
	util.PanicErr(err)

	// Close if either peer wants
	conWanted, err := verifyConnWanted(pc, helloMsg)
	util.PanicErr(err)
	if !conWanted {
		// TODO: signal peer dead on bus
		return
	}

	// Does PeerRoutine start with different conditions?
	go PeerRoutine(pc, helloMsg)
}

func PeerRoutine(pc util.PeerConn, data HelloMessage) {
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
