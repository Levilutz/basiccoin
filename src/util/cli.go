package util

import "flag"

type CLIArgs struct {
	LocalAddr string `json:"localAddr"`
	SeedAddr  string `json:"seedAddr"`
}

func ParseCLIArgs() CLIArgs {
	// Define args
	localAddr := flag.String("addr", "", "Address to host from")
	seedAddr := flag.String("seed", "", "Seed peer, or nothing to create new network")

	// Do the parse
	flag.Parse()

	// Insert into Constants
	if *localAddr != "" {
		Constants.LocalAddr = *localAddr
	}
	if *seedAddr != "" {
		Constants.SeedAddr = *seedAddr
	}

	// Return all (even those in constants)
	return CLIArgs{
		LocalAddr: *localAddr,
		SeedAddr:  *seedAddr,
	}
}
