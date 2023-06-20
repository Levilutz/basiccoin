package main

import (
	"github.com/levilutz/basiccoin/internal/peerfactory"
	"github.com/levilutz/basiccoin/internal/pubsub"
	"github.com/levilutz/basiccoin/src/kern"
)

func main() {
	runtimeId := kern.NewHashTRand().String()

	// Create params for each component
	var peerFactoryParams peerfactory.Params
	if false {
		peerFactoryParams = peerfactory.ProdParams(true, runtimeId)
	} else {
		peerFactoryParams = peerfactory.DevParams(true, runtimeId)
	}

	// Make the event bus
	pubSub := pubsub.NewPubSub()

	// Create app components
	peerFactory := peerfactory.NewPeerFactory(peerFactoryParams, pubSub)

	// Start app components
	peerFactory.Loop()
}
