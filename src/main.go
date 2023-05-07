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

const Version = "0.1.0"

type VersionResp struct {
	Version     string `json:"version"`
	CurrentTime int64  `json:"currentTime"`
}

type PeerRecord struct {
	LastUpdated time.Time
	Version     VersionResp
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

func (pc *PeersContainer) Upsert(addr string, resp VersionResp) {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	pc.peers[addr] = PeerRecord{time.Now(), resp}
}

func (pc *PeersContainer) Get(addr string) (record PeerRecord) {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	record = pc.peers[addr]
	return
}

func (pc *PeersContainer) Print() {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	for addr, record := range pc.peers {
		fmt.Printf("%s\t%d\t%v\n", addr, record.LastUpdated.Unix(), record.Version)
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
		resp, err := utils.RetryGetBody[VersionResp]("http://"+*seedAddr+"/version", 3)
		if err != nil {
			fmt.Printf("Failed to connect to seed peer: %s", err)
			os.Exit(1)
		}
		peers.Upsert(*seedAddr, *resp)
	}
	peers.Print()

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
