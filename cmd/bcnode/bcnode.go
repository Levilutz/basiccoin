package main

import (
	"github.com/levilutz/basiccoin/internal/peerfactory"
	"github.com/levilutz/basiccoin/internal/pubsub"
)

func main() {
	// Create params for each component
	var peerFactoryParams peerfactory.Params
	if false {
		peerFactoryParams = peerfactory.ProdParams(true, "")
	} else {
		peerFactoryParams = peerfactory.DevParams(true, "")
	}

	// Make the event bus
	pubSub := pubsub.NewPubSub()

	// Create app components
	peerFactory := peerfactory.NewPeerFactory(peerFactoryParams, pubSub)

	// Start app components
	peerFactory.Loop()
}
