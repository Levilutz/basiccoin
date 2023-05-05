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
)

const Version = "0.1.0"

type VersionResp struct {
	Version     string `json:"version"`
	CurrentTime int64  `json:"currentTime"`
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

func requestPeerVersion(peerAddr string) VersionResp {
	resp, _ := http.Get("http://" + peerAddr + "/version")
	body, _ := io.ReadAll(resp.Body)
	var content VersionResp
	json.Unmarshal(body, &content)
	return content
}

func main() {
	localAddr, seedAddr := getCLIArgs()

	neighbors := make(map[string]VersionResp)
	if *seedAddr != "" {
		neighbors[*seedAddr] = requestPeerVersion(*seedAddr)
	}
	fmt.Println(neighbors)

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
