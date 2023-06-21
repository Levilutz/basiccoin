package main

import (
	"time"

	"github.com/levilutz/basiccoin/internal/peerfactory"
	"github.com/levilutz/basiccoin/internal/pubsub"
)

func main() {
	flags := ParseFlags()

	// Create params for each component
	var peerFactoryParams peerfactory.Params
	var printUpdateFreq time.Duration
	if flags.Dev {
		peerFactoryParams = peerfactory.DevParams(flags.Listen, flags.LocalAddr)
		printUpdateFreq = time.Second * 5
	} else {
		peerFactoryParams = peerfactory.ProdParams(flags.Listen, flags.LocalAddr)
		printUpdateFreq = time.Second * 60
	}

	// Make the event bus
	pubSub := pubsub.NewPubSub()

	// Create app components
	peerFactory := peerfactory.NewPeerFactory(peerFactoryParams, pubSub)

	// Set seed peer
	if flags.SeedAddr != "" {
		peerFactory.SetSeed(flags.SeedAddr)
	}

	// Start app components
	go peerFactory.Loop()

	// Trigger updates forever
	for {
		pubSub.PrintUpdate.Pub(pubsub.PrintUpdateEvent{
			Peer:        false,
			PeerFactory: true,
		})
		time.Sleep(printUpdateFreq)
	}
}
