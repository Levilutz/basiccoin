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

	// Start app components
	go peerFactory.Loop()

	// Initialize connection to seed peer
	if flags.SeedAddr != "" {
		pubSub.PeersReceived.Pub(pubsub.PeersReceivedEvent{
			PeerAddrs: map[string]string{
				"": flags.SeedAddr,
			},
		})
	}

	// Trigger updates forever
	for {
		pubSub.PrintUpdate.Pub(pubsub.PrintUpdateEvent{
			Peer:        flags.Dev,
			PeerFactory: true,
		})
		time.Sleep(printUpdateFreq)
	}
}
