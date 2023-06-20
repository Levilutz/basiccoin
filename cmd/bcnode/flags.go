package main

import "flag"

type Flags struct {
	Dev       bool
	Listen    bool
	LocalAddr string
	SeedAddr  string
}

func ParseFlags() Flags {
	// Parse from command line
	dev := flag.Bool("dev", false, "Whether to start the server in dev mode")
	listen := flag.Bool("listen", true, "Whether to listen for inbound connections")
	localAddr := flag.String("addr", "", "Local address to host from")
	seedAddr := flag.String("seed", "", "Seed peer")
	flag.Parse()

	// Fill in other defaults
	if *localAddr == "" && *seedAddr == "" && !*dev {
		*seedAddr = "coin.levilutz.com:21720"
	}

	return Flags{
		Dev:       *dev,
		Listen:    *listen,
		LocalAddr: *localAddr,
		SeedAddr:  *seedAddr,
	}
}
