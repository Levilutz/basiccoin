package main

import (
	"time"

	"github.com/levilutz/basiccoin/internal/inv"
	"github.com/levilutz/basiccoin/internal/peerfactory"
	"github.com/levilutz/basiccoin/internal/pubsub"
	"github.com/levilutz/basiccoin/pkg/core"
)

func main() {
	flags := ParseFlags()

	// Create params for each component
	var coreParams core.Params
	var peerFactoryParams peerfactory.Params
	var printUpdateFreq time.Duration
	if flags.Dev {
		coreParams = core.DevNetParams()
		peerFactoryParams = peerfactory.DevParams(flags.Listen, flags.LocalAddr)
		printUpdateFreq = time.Second * 5
	} else {
		coreParams = core.ProdNetParams()
		peerFactoryParams = peerfactory.ProdParams(flags.Listen, flags.LocalAddr)
		printUpdateFreq = time.Second * 60
	}

	// Make the event bus and shared inventory
	pubSub := pubsub.NewPubSub()
	inv := inv.NewInv(coreParams)

	// Create app components
	peerFactory := peerfactory.NewPeerFactory(peerFactoryParams, pubSub, inv)

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
