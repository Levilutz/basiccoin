package main

import (
	"flag"

	"github.com/levilutz/basiccoin/pkg/core"
)

type Flags struct {
	Dev         bool
	Listen      bool
	LocalAddr   string
	SeedAddr    string
	Miners      int
	PayoutPkh   core.HashT
	HttpEnabled bool
	HttpPort    int
}

func ParseFlags() Flags {
	// Parse from command line
	dev := flag.Bool("dev", false, "Whether to start the server in dev mode")
	listen := flag.Bool("listen", true, "Whether to listen for inbound connections")
	localAddr := flag.String("addr", "", "Local address to host from")
	seedAddr := flag.String("seed", "", "Seed peer")
	miners := flag.Int("miners", 0, "Number of threads to mine with")
	payoutPkh := flag.String("payout", "", "Public key hash to pay out miner reward to")
	httpPort := flag.Int("http", 0, "Port to enable and host the http server from")

	flag.Parse()

	// Validate and convert types
	var payoutPkhHash core.HashT
	if *miners > 0 {
		if *payoutPkh == "" {
			panic("Must set payout pkh when mining")
		}
		payoutPkhHash = core.NewHashTFromStringAssert(*payoutPkh)
	} else {
		payoutPkhHash = core.HashT{}
	}

	// Fill in other defaults
	if *localAddr == "" && *seedAddr == "" && !*dev {
		*seedAddr = "coin.levilutz.com:21720"
	}

	return Flags{
		Dev:         *dev,
		Listen:      *listen,
		LocalAddr:   *localAddr,
		SeedAddr:    *seedAddr,
		Miners:      *miners,
		PayoutPkh:   payoutPkhHash,
		HttpEnabled: *httpPort != 0,
		HttpPort:    *httpPort,
	}
}
