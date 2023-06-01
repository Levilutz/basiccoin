package util

import "flag"

type CLIArgs struct {
	Listen    bool   `json:"listen"`
	LocalAddr string `json:"localAddr"`
	Miners    int    `json:"miners"`
	SeedAddr  string `json:"seedAddr"`
	Verbose   bool   `json:"verbose"`
}

func ParseCLIArgs() CLIArgs {
	// Define args
	listen := flag.Bool("listen", true, "Whether to listen for inbound connections")
	localAddr := flag.String("addr", "", "Address to host from")
	miners := flag.Int("miners", 0, "How many miner instances (defaults to 0)")
	seedAddr := flag.String("seed", "", "Seed peer, or nothing to create new network")
	verbose := flag.Bool("v", false, "Whether to show debug logs")

	// Do the parse
	flag.Parse()

	// Validate
	if *localAddr == "" && *seedAddr == "" {
		panic("Must provide either --addr or --seed")
	}

	// Insert into Constants
	Constants.Listen = *listen
	Constants.Miners = *miners
	if *localAddr != "" {
		Constants.LocalAddr = *localAddr
	}
	if *seedAddr != "" {
		Constants.SeedAddr = *seedAddr
	}
	if *verbose {
		Constants.Debug = true
	}

	// Return all (even those in constants)
	return CLIArgs{
		Listen:    *listen,
		Miners:    *miners,
		LocalAddr: *localAddr,
		SeedAddr:  *seedAddr,
	}
}
