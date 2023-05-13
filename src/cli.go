package main

import "flag"

type CLIArgs struct {
	SeedAddr string `json:"seedAddr"`
}

func GetCLIArgs() CLIArgs {
	seedAddr := flag.String("seed", "", "Seed peer, or nothing to create new network")

	flag.Parse()
	return CLIArgs{
		SeedAddr: *seedAddr,
	}
}
