package rest

import (
	"fmt"
	"net/http"

	"github.com/levilutz/basiccoin/src/db"
	"github.com/levilutz/basiccoin/src/util"
)

type MainQueryHandler interface {
	HandleBalanceQuery(rCh chan<- uint64, publicKeyHash db.HashT)
}

func Start(m MainQueryHandler) {
	handler := Handler{m: m}

	http.HandleFunc("/ping", handler.handleGetPing)
	http.HandleFunc("/balance", handler.handleBalance)
	http.HandleFunc("/tx", handler.handleTx)

	portStr := fmt.Sprintf(":%d", util.Constants.HttpPort)
	http.ListenAndServe(portStr, nil)
}
