package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/levilutz/basiccoin/src/p2p"
	"github.com/levilutz/basiccoin/src/utils"
)

const (
	allowedFailures = 3
	pollingPeriod   = 5
)

type VersionResp struct {
	Version     string `json:"version"`
	CurrentTime int64  `json:"currentTime"`
}

func updatePeerVersion(peers *p2p.Peers, addr string) error {
	resp, midTimeMicro, err := utils.RetryGetBody[VersionResp]("http://"+addr+"/version", 3)
	if err != nil {
		totalFailures := peers.IncrementFailures(addr)
		if totalFailures > allowedFailures {
			peers.DropPeer(addr)
		}
		return err
	}
	peers.Upsert(addr, &p2p.PeerData{
		Version:         resp.Version,
		TimeOffsetMicro: resp.CurrentTime - midTimeMicro,
	})
	return nil
}

func updatePeerLoop(peers *p2p.Peers, interval int, kill <-chan bool) {
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	for {
		select {
		case <-kill:
			return
		case <-ticker.C:
			addrs := peers.GetAddrs()
			if len(addrs) == 0 {
				fmt.Println("All peers lost")
			}
			for _, addr := range addrs {
				addr := addr
				go func() {
					updatePeerVersion(peers, addr)
					fmt.Println("Peers:")
					peers.Print()
				}()
			}
		}
	}
}

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

	peers := p2p.NewPeers()

	if *seedAddr != "" {
		err := updatePeerVersion(peers, *seedAddr)
		if err != nil {
			fmt.Printf("Failed to connect to seed peer: %s", err)
			os.Exit(1)
		}
	}
	peers.Print()
	go updatePeerLoop(peers, pollingPeriod, nil)

	r := gin.Default()
	r.GET("/ping", getPing)
	p2p.Mount(r)

	fmt.Printf("Starting at %s\n", *localAddr)
	r.SetTrustedProxies(nil)
	r.Run(*localAddr)
}
