package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/levilutz/basiccoin/pkg/core"
)

type Flags struct {
	Dev               bool
	Listen            bool
	LocalAddr         string
	SeedAddrs         []string
	Miners            int
	PayoutPkh         core.HashT
	HttpPort          int
	HttpAdminEnabled  bool
	HttpWalletEnabled bool
	HttpAdminPw       string
	SaveDir           *string
}

func ParseFlags() Flags {
	// Parse from command line
	dev := flag.Bool("dev", false, "Whether to start the server in dev mode")
	listen := flag.Bool("listen", false, "Whether to listen for inbound connections")
	newNetwork := flag.Bool("new-network", false, "Whether to start a new network (if true, 'seeds' is ignored)")
	localAddr := flag.String("addr", "", "Local address to host from")
	seedAddrs := flag.String("seeds", "", "Seed peers, comma-separated")
	miners := flag.Int("miners", 0, "Number of threads to mine with")
	payoutPkh := flag.String("payout", "", "Public key hash to pay out miner reward to")
	httpPort := flag.Int("http", 80, "Port to host the http server from")
	httpAdmin := flag.Bool("http-admin", false, "Whether to enable the admin http server")
	httpWallet := flag.Bool("http-wallet", false, "Whether to enable the wallet http server")
	httpAdminPw := flag.String("admin-pw", "", "Password for the admin http endpoints")
	saveDir := flag.String("save-dir", "", "Directory to save the chain")

	flag.Parse()

	// Validate, convert types, fill in other defaults
	var payoutPkhHash core.HashT
	if *miners > 0 {
		if *payoutPkh == "" {
			panic("Must set payout pkh when mining")
		}
		payoutPkhHash = core.NewHashTFromStringAssert(*payoutPkh)
	} else {
		payoutPkhHash = core.HashT{}
	}

	var seedAddrsList []string
	if *newNetwork || (*seedAddrs == "" && *dev) {
		seedAddrsList = []string{}
	} else if *seedAddrs != "" {
		seedAddrsList = strings.Split(*seedAddrs, ",")
	} else {
		seedAddrsList = []string{
			"coin1.levilutz.com:21720",
			"coin2.levilutz.com:21720",
			"coin3.levilutz.com:21720",
		}
	}

	var saveDirReal *string
	if *saveDir != "" {
		if err := os.MkdirAll(*saveDir, 0750); err != nil {
			panic(fmt.Sprintf("failed to create save dir: %s", err))
		}
		saveDirReal = saveDir
	} else {
		saveDirReal = nil
	}

	return Flags{
		Dev:               *dev,
		Listen:            *listen,
		LocalAddr:         *localAddr,
		SeedAddrs:         seedAddrsList,
		Miners:            *miners,
		PayoutPkh:         payoutPkhHash,
		HttpPort:          *httpPort,
		HttpAdminEnabled:  *httpAdmin,
		HttpWalletEnabled: *httpWallet,
		HttpAdminPw:       *httpAdminPw,
		SaveDir:           saveDirReal,
	}
}
