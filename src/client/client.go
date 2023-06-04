package client

import (
	"fmt"
	"io"
	"net/http"

	"github.com/levilutz/basiccoin/src/util"
)

type MainQueryHandler interface {
	HandlePingQuery(rCh chan<- string)
}

func Start(m MainQueryHandler) {
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		rCh := make(chan string)
		m.HandlePingQuery(rCh)
		resp := <-rCh
		io.WriteString(w, resp)
	})

	portStr := fmt.Sprintf(":%d", util.Constants.HttpPort)
	http.ListenAndServe(portStr, nil)
}
