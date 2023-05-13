package main

import (
	"bufio"
	"errors"
	"fmt"
	"net"

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
func verifyConnWanted(
	r *bufio.Reader, w *bufio.Writer, helloMsg HelloMessage,
) (bool, error) {
	// Decide if we want to continue and tell them
	var err error
	shouldConn := shouldConnect(helloMsg)
	if shouldConn {
		err = TransmitSimpleMessage(w, "continue")
	} else {
		err = TransmitSimpleMessage(w, "close")
	}
	if err != nil {
		return false, err
	}

	// Receive whether they want to continue
	contMsg, err := util.RetryReadLine(r, 8)
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

func GreetPeer(c *net.TCPConn) error {
	defer func() {
		if r := recover(); r != nil {
			// TODO: signal peer dead on bus
			fmt.Printf("Failed GreetPeer: %v\n", r)
		}
	}()

	r := bufio.NewReader(c)
	w := bufio.NewWriter(c)

	// Hello handshake
	err := NewHelloMessage().Transmit(w)
	if err != nil {
		// TODO: signal peer dead on bus
		return err
	}
	err = ConsumeExpected(r, "ack:hello")
	if err != nil {
		// TODO: signal peer dead on bus
		return err
	}
	err = ConsumeExpected(r, "hello")
	if err != nil {
		// TODO: signal peer dead on bus
		return err
	}
	helloMsg, err := ReceiveHelloMessage(r)
	if err != nil {
		// TODO: signal peer dead on bus
		return err
	}
	err = TransmitSimpleMessage(w, "ack:hello")
	if err != nil {
		// TODO: signal peer dead on bus
		return err
	}

	// Close if either peer wants
	conWanted, err := verifyConnWanted(r, w, helloMsg)
	if err != nil {
		// TODO: signal peer dead on bus
		return err
	}
	if !conWanted {
		// TODO: signal peer dead on bus
		return errors.New("peer does not want connection")
	}

	go PeerRoutine(r, w, helloMsg)
	return nil
}

func ReceivePeerGreeting(c *net.TCPConn) {
	defer func() {
		if r := recover(); r != nil {
			// TODO: signal peer dead on bus
			fmt.Printf("Failed ReceivePeerGreeting: %v\n", r)
		}
	}()

	r := bufio.NewReader(c)
	w := bufio.NewWriter(c)

	// Hello handshake
	err := ConsumeExpected(r, "hello")
	util.PanicErr(err)
	helloMsg, err := ReceiveHelloMessage(r)
	util.PanicErr(err)
	err = TransmitSimpleMessage(w, "ack:hello")
	util.PanicErr(err)
	err = NewHelloMessage().Transmit(w)
	util.PanicErr(err)
	err = ConsumeExpected(r, "ack:hello")
	util.PanicErr(err)

	// Close if either peer wants
	conWanted, err := verifyConnWanted(r, w, helloMsg)
	util.PanicErr(err)
	if !conWanted {
		// TODO: signal peer dead on bus
		return
	}

	// Does PeerRoutine start with different conditions?
	go PeerRoutine(r, w, helloMsg)
}

func PeerRoutine(r *bufio.Reader, w *bufio.Writer, data HelloMessage) {
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
