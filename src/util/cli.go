package util

import "flag"

type CLIArgs struct {
	HttpPort  int    `json:"httpPort"`
	Listen    bool   `json:"listen"`
	LocalAddr string `json:"localAddr"`
	Miners    int    `json:"miners"`
	SeedAddr  string `json:"seedAddr"`
	Verbosity int    `json:"verbosity"`
}

func ParseCLIArgs() CLIArgs {
	// Define args
	httpPort := flag.Int("httpPort", -1, "Port to listen for http requests on (-1 to not listen)")
	listen := flag.Bool("listen", true, "Whether to listen for inbound connections")
	localAddr := flag.String("addr", "", "Address to host from")
	miners := flag.Int("miners", 0, "How many miner instances (defaults to 0)")
	seedAddr := flag.String("seed", "", "Seed peer, or nothing to create new network")
	verbose1 := flag.Bool("v", false, "Whether to show debug logs")
	verbose2 := flag.Bool("vv", false, "Whether to show more debug logs")

	// Do the parse
	flag.Parse()

	// Validate
	if *localAddr == "" && *seedAddr == "" {
		panic("Must provide either --addr or --seed")
	}

	// Insert into Constants
	Constants.HttpPort = *httpPort
	Constants.Listen = *listen
	Constants.Miners = *miners
	if *localAddr != "" {
		Constants.LocalAddr = *localAddr
	}
	if *seedAddr != "" {
		Constants.SeedAddr = *seedAddr
	}
	if *verbose2 {
		Constants.DebugLevel = 2
	} else if *verbose1 {
		Constants.DebugLevel = 1
	} else {
		Constants.DebugLevel = 0
	}

	// Return all (even those in constants)
	return CLIArgs{
		Listen:    *listen,
		Miners:    *miners,
		LocalAddr: *localAddr,
		SeedAddr:  *seedAddr,
		Verbosity: Constants.DebugLevel,
	}
}
