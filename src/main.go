package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/levilutz/basiccoin/src/p2p"
	"github.com/levilutz/basiccoin/src/utils"
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
	utils.Constants.LocalAddr = *localAddr

	pn := p2p.NewP2pNetwork()

	if *seedAddr != "" {
		err := pn.RetryAddPeer(*seedAddr, true)
		if err != nil {
			fmt.Printf("failed to connect to seed peer: %s", err)
			os.Exit(1)
		}
	}
	pn.Print()
	go pn.SyncLoop(true, nil)

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.GET("/ping", getPing)
	p2p.Mount(r, pn)

	fmt.Printf("starting at %s\n", *localAddr)
	r.SetTrustedProxies(nil)
	r.Run(*localAddr)
}
