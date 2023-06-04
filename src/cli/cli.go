package main

import (
	"flag"
	"fmt"

	"github.com/levilutz/basiccoin/src/util"
)

func ParseCLIArgs() {
	// Define args
	verbose1 := flag.Bool("v", false, "Whether to show debug logs")
	verbose2 := flag.Bool("vv", false, "Whether to show more debug logs")

	// Do the parse
	flag.Parse()

	// Insert into constants
	if *verbose2 {
		util.Constants.DebugLevel = 2
	} else if *verbose1 {
		util.Constants.DebugLevel = 1
	} else {
		util.Constants.DebugLevel = 0
	}

	fmt.Println(util.Constants.DebugLevel)
}
