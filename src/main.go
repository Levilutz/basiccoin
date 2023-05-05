package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
)

const Version = "0.1.0"

func getCLIArgs() (string, *string) {
	if len(os.Args) == 1 {
		return "21720", nil
	} else if len(os.Args) == 2 {
		return os.Args[1], nil
	}
	return os.Args[1], &os.Args[2]
}

func getPing(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "pong")
}

func getVersion(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, Version)
}

func main() {
	selfPort, seedPartnerAddr := getCLIArgs()

	neighbors := make([]string, 1)
	if seedPartnerAddr != nil {
		neighbors[0] = *seedPartnerAddr
	}

	http.HandleFunc("/ping", getPing)
	http.HandleFunc("/version", getVersion)

	fmt.Printf("Starting at 0.0.0.0:%s\n", selfPort)
	err := http.ListenAndServe("0.0.0.0:"+selfPort, nil)

	if errors.Is(err, http.ErrServerClosed) {
		fmt.Println("Server closed")
	} else if err != nil {
		fmt.Fprintf(os.Stderr, "error starting server: %v\n", err)
		os.Exit(1)
	}
}
