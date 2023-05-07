package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

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

type PeerRecord struct {
	LastSuccessfullyUpdated time.Time
	ConnectionFailures      int
	Version                 *VersionResp
}

type PeersContainer struct {
	mu    sync.Mutex
	peers map[string]PeerRecord
}

func NewPeersContainer() *PeersContainer {
	return &PeersContainer{
		peers: make(map[string]PeerRecord),
	}
}

func (pc *PeersContainer) Upsert(addr string, resp *VersionResp) {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	pc.peers[addr] = PeerRecord{time.Now(), 0, resp}
}

func (pc *PeersContainer) IncrementFailures(addr string) (totalFailures int) {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	if record, ok := pc.peers[addr]; ok {
		record.ConnectionFailures++
		totalFailures = record.ConnectionFailures
		pc.peers[addr] = record
	} else {
		pc.peers[addr] = PeerRecord{time.Time{}, 1, nil}
		totalFailures = 1
	}
	return
}

func (pc *PeersContainer) Get(addr string) (record PeerRecord) {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	record = pc.peers[addr]
	return
}

func (pc *PeersContainer) GetAddrs() (addrs []string) {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	addrs = make([]string, len(pc.peers))
	i := 0
	for k := range pc.peers {
		addrs[i] = k
		i++
	}
	return
}

func (pc *PeersContainer) DropPeer(addr string) {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	delete(pc.peers, addr)
}

func (pc *PeersContainer) Print() {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	for addr, record := range pc.peers {
		fmt.Printf(
			"%s\t%d\t%v\n",
			addr,
			record.LastSuccessfullyUpdated.Unix(),
			record.Version,
		)
	}
}

func updatePeerVersion(peers *PeersContainer, addr string) error {
	resp, err := utils.RetryGetBody[VersionResp]("http://"+addr+"/version", 3)
	if err != nil {
		totalFailures := peers.IncrementFailures(addr)
		if totalFailures > allowedFailures {
			peers.DropPeer(addr)
		}
		return err
	}
	peers.Upsert(addr, resp)
	return nil
}

func updatePeerLoop(peers *PeersContainer, interval int, kill <-chan bool) {
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

	peers := NewPeersContainer()

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
