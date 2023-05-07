package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/levilutz/basiccoin/src/p2p"
)

func getCLIArgs() (localAddr, seedAddr *string) {
	localAddr = flag.String(
		"localAddr", "0.0.0.0:21720", "Local address to host server",
	)
	seedAddr = flag.String(
		"seedAddr", "", "Seed partner, or nothing to create new network",
	)
	flag.Parse()
	return
}

func getPing(c *gin.Context) {
	c.String(http.StatusOK, "pong")
}

func main() {
	localAddr, seedAddr := getCLIArgs()

	pn := p2p.NewP2pNetwork()

	if *seedAddr != "" {
		err := pn.RetryAddPeer(*seedAddr)
		if err != nil {
			fmt.Printf("Failed to connect to seed peer: %s", err)
			os.Exit(1)
		}
	}
	pn.Print()
	go pn.SyncLoop(true, nil)

	r := gin.Default()
	r.GET("/ping", getPing)
	p2p.Mount(r)

	fmt.Printf("Starting at %s\n", *localAddr)
	r.SetTrustedProxies(nil)
	r.Run(*localAddr)
}
