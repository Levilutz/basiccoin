package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/levilutz/basiccoin/src/db"
	"github.com/levilutz/basiccoin/src/util"
)

type MainQueryHandler interface {
	HandleBalanceQuery(rCh chan<- uint64, publicKeyHash db.HashT)
}

func Start(m MainQueryHandler) {
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "pong")
	})

	http.HandleFunc("/balance", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.WriteHeader(405)
			io.WriteString(w, "method not allowed: "+r.Method)
			return
		}
		publicKeyHashes, ok := r.URL.Query()["publicKeyHash"]
		if !ok || len(publicKeyHashes) < 1 {
			w.WriteHeader(400)
			io.WriteString(w, "must provide public key hash")
			return
		}
		pkh, err := db.StringToHash(publicKeyHashes[0])
		if err != nil {
			w.WriteHeader(400)
			io.WriteString(w, err.Error())
			return
		}
		rCh := make(chan uint64)
		m.HandleBalanceQuery(rCh, pkh)
		resp := <-rCh
		io.WriteString(w, strconv.FormatUint(resp, 10))
	})

	http.HandleFunc("/tx", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(405)
			io.WriteString(w, "method not allowed: "+r.Method)
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(400)
			io.WriteString(w, "failed to read body: "+err.Error())
			return
		}
		tx := db.Tx{}
		if err = json.Unmarshal(body, &tx); err != nil {
			w.WriteHeader(422)
			io.WriteString(w, "failed to parse json: "+err.Error())
			return
		}
		if !tx.HasSurplus() {
			w.WriteHeader(400)
			io.WriteString(w, "tx without surplus would never be included")
			return
		}
	})

	portStr := fmt.Sprintf(":%d", util.Constants.HttpPort)
	http.ListenAndServe(portStr, nil)
}
