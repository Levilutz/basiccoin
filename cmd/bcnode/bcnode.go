package main

import (
	"time"

	"github.com/levilutz/basiccoin/internal/chain"
	"github.com/levilutz/basiccoin/internal/inv"
	"github.com/levilutz/basiccoin/internal/miner"
	"github.com/levilutz/basiccoin/internal/peerfactory"
	"github.com/levilutz/basiccoin/internal/pubsub"
	"github.com/levilutz/basiccoin/pkg/core"
)

func main() {
	flags := ParseFlags()

	// Create params for each component
	var coreParams core.Params
	var minerParams miner.Params
	var peerFactoryParams peerfactory.Params
	var printUpdateFreq time.Duration
	if flags.Dev {
		coreParams = core.DevNetParams()
		minerParams = miner.NewParams(flags.PayoutPkh)
		peerFactoryParams = peerfactory.DevParams(flags.Listen, flags.LocalAddr)
		printUpdateFreq = time.Second * 5
	} else {
		coreParams = core.ProdNetParams()
		minerParams = miner.NewParams(flags.PayoutPkh)
		peerFactoryParams = peerfactory.ProdParams(flags.Listen, flags.LocalAddr)
		printUpdateFreq = time.Second * 60
	}

	// Make the event bus and shared inventory
	pubSub := pubsub.NewPubSub()
	inv := inv.NewInv(coreParams)

	// Create app components
	chain := chain.NewChain(pubSub, inv, flags.Miners > 0)
	peerFactory := peerfactory.NewPeerFactory(peerFactoryParams, pubSub, inv)
	miners := make([]*miner.Miner, flags.Miners)
	for i := 0; i < flags.Miners; i++ {
		miners[i] = miner.NewMiner(minerParams, pubSub, inv)
	}

	// Set seed peer
	if flags.SeedAddr != "" {
		peerFactory.SetSeed(flags.SeedAddr)
	}

	// Start app components (order matters)
	for i := 0; i < flags.Miners; i++ {
		go miners[i].Loop()
	}
	time.Sleep(time.Millisecond * 250)
	go chain.Loop()
	time.Sleep(time.Millisecond * 250)
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
