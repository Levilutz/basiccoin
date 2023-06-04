package client

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/levilutz/basiccoin/src/db"
	"github.com/levilutz/basiccoin/src/util"
)

type MainQueryHandler interface {
	HandlePingQuery(rCh chan<- string)
	HandleBalanceQuery(rCh chan<- uint64, publicKeyHash db.HashT)
}

func Start(m MainQueryHandler) {
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		rCh := make(chan string)
		m.HandlePingQuery(rCh)
		resp := <-rCh
		io.WriteString(w, resp)
	})

	http.HandleFunc("/balance", func(w http.ResponseWriter, r *http.Request) {
		publicKeyHashes, ok := r.URL.Query()["publicKeyHash"]
		if !ok || len(publicKeyHashes) < 1 {
			w.WriteHeader(404)
			io.WriteString(w, "not found")
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

	portStr := fmt.Sprintf(":%d", util.Constants.HttpPort)
	http.ListenAndServe(portStr, nil)
}
