package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/levilutz/basiccoin/src/p2p"
	"github.com/levilutz/basiccoin/src/utils"
)

const (
	Version         = "0.1.0"
	allowedFailures = 3
	pollingPeriod   = 5
)

type VersionResp struct {
	Version     string `json:"version"`
	CurrentTime int64  `json:"currentTime"`
}

func updatePeerVersion(peers *p2p.Peers, addr string) error {
	sentTime := time.Now().UnixMicro()
	resp, err := utils.RetryGetBody[VersionResp]("http://"+addr+"/version", 3)
	respTime := time.Now().UnixMicro()
	// TODO: This is biased by json unmarshalling in RetryGetBody - do inside instead
	midTime := (sentTime + respTime) / 2
	if err != nil {
		totalFailures := peers.IncrementFailures(addr)
		if totalFailures > allowedFailures {
			peers.DropPeer(addr)
		}
		return err
	}
	peers.Upsert(addr, &p2p.PeerData{
		Version:         resp.Version,
		TimeOffsetMicro: resp.CurrentTime - midTime,
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
				go updatePeerVersion(peers, addr)
				peers.Print()
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

func getPing(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "pong")
}

func getVersion(w http.ResponseWriter, r *http.Request) {
	b, _ := json.Marshal(VersionResp{
		Version,
		time.Now().UnixMicro(),
	})
	w.Write(b)
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

	http.HandleFunc("/ping", getPing)
	http.HandleFunc("/version", getVersion)

	fmt.Printf("Starting at %s\n", *localAddr)
	err := http.ListenAndServe(*localAddr, nil)

	if errors.Is(err, http.ErrServerClosed) {
		fmt.Println("Server closed")
	} else if err != nil {
		fmt.Fprintf(os.Stderr, "error starting server: %v\n", err)
		os.Exit(1)
	}
}
