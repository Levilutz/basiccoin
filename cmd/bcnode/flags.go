package main

import (
	"flag"

	"github.com/levilutz/basiccoin/pkg/core"
)

type Flags struct {
	Dev               bool
	Listen            bool
	LocalAddr         string
	SeedAddr          string
	Miners            int
	PayoutPkh         core.HashT
	HttpPort          int
	HttpAdminEnabled  bool
	HttpWalletEnabled bool
	HttpAdminPw       string
}

func ParseFlags() Flags {
	// Parse from command line
	dev := flag.Bool("dev", false, "Whether to start the server in dev mode")
	listen := flag.Bool("listen", true, "Whether to listen for inbound connections")
	localAddr := flag.String("addr", "", "Local address to host from")
	seedAddr := flag.String("seed", "", "Seed peer")
	miners := flag.Int("miners", 0, "Number of threads to mine with")
	payoutPkh := flag.String("payout", "", "Public key hash to pay out miner reward to")
	httpPort := flag.Int("http", 80, "Port to host the http server from")
	httpAdmin := flag.Bool("http-admin", false, "Whether to enable the admin http server")
	httpWallet := flag.Bool("http-wallet", false, "Whether to enable the wallet http server")
	httpAdminPw := flag.String("admin-pw", "", "Password for the admin http endpoints")

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
		Dev:               *dev,
		Listen:            *listen,
		LocalAddr:         *localAddr,
		SeedAddr:          *seedAddr,
		Miners:            *miners,
		PayoutPkh:         payoutPkhHash,
		HttpPort:          *httpPort,
		HttpAdminEnabled:  *httpAdmin,
		HttpWalletEnabled: *httpWallet,
		HttpAdminPw:       *httpAdminPw,
	}
}
