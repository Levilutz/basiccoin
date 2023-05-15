package util

import "flag"

type CLIArgs struct {
	Listen    bool   `json:"listen"`
	LocalAddr string `json:"localAddr"`
	SeedAddr  string `json:"seedAddr"`
}

func ParseCLIArgs() CLIArgs {
	// Define args
	listen := flag.Bool("listen", true, "Whether to listen for inbound connections")
	localAddr := flag.String("addr", "", "Address to host from")
	seedAddr := flag.String("seed", "", "Seed peer, or nothing to create new network")

	// Do the parse
	flag.Parse()

	// Insert into Constants
	Constants.Listen = *listen
	if *localAddr != "" {
		Constants.LocalAddr = *localAddr
	}
	if *seedAddr != "" {
		Constants.SeedAddr = *seedAddr
	}

	// Return all (even those in constants)
	return CLIArgs{
		Listen:    *listen,
		LocalAddr: *localAddr,
		SeedAddr:  *seedAddr,
	}
}
