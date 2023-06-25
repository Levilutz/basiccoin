package main

import (
	"time"

	"github.com/levilutz/basiccoin/internal/bus"
	"github.com/levilutz/basiccoin/internal/chain"
	"github.com/levilutz/basiccoin/internal/inv"
	"github.com/levilutz/basiccoin/internal/miner"
	"github.com/levilutz/basiccoin/internal/peerfactory"
	"github.com/levilutz/basiccoin/internal/rest"
	"github.com/levilutz/basiccoin/pkg/core"
)

func main() {
	flags := ParseFlags()

	// Create params for each component
	var coreParams core.Params
	var minerParams miner.Params
	var peerFactoryParams peerfactory.Params
	var restParams rest.Params
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
	if flags.HttpAdminEnabled || flags.HttpWalletEnabled {
		restParams = rest.NewParams(
			flags.HttpPort,
			flags.HttpAdminEnabled,
			flags.HttpWalletEnabled,
			flags.HttpAdminPw,
			flags.Dev,
		)
	}

	// Make the event bus and shared inventory
	msgBus := bus.NewBus()
	inv := inv.NewInv(coreParams)

	// Create app components
	chain := chain.NewChain(msgBus, inv, flags.Miners > 0)
	peerFactory := peerfactory.NewPeerFactory(peerFactoryParams, msgBus, inv)
	miners := make([]*miner.Miner, flags.Miners)
	for i := 0; i < flags.Miners; i++ {
		miners[i] = miner.NewMiner(minerParams, msgBus, inv)
	}
	var restServer *rest.Server
	if flags.HttpAdminEnabled || flags.HttpWalletEnabled {
		restServer = rest.NewServer(restParams, msgBus, inv)
	}

	// Set seed peer
	if len(flags.SeedAddrs) > 0 {
		peerFactory.SetSeeds(flags.SeedAddrs)
	}

	// Start app components (order matters)
	for i := 0; i < flags.Miners; i++ {
		go miners[i].Loop()
	}
	time.Sleep(time.Millisecond * 250)
	go chain.Loop()
	time.Sleep(time.Millisecond * 250)
	go peerFactory.Loop()
	if flags.HttpAdminEnabled || flags.HttpWalletEnabled {
		go restServer.Start()
	}

	// Trigger updates and watch for terminate command
	printUpdatesTicker := time.NewTicker(printUpdateFreq)
	terminateSubCh := msgBus.Terminate.SubCh()
	for {
		select {
		case <-printUpdatesTicker.C:
			msgBus.PrintUpdate.Pub(bus.PrintUpdateEvent{
				Peer:        false,
				PeerFactory: true,
			})

		case <-terminateSubCh.C:
			panic("terminated")
		}
	}
}
