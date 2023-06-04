package rest

import (
	"fmt"
	"net/http"

	"github.com/levilutz/basiccoin/src/util"
)

func Start(m MainQueryHandler) {
	handler := Handler{m: m}

	http.HandleFunc("/ping", handler.handleGetPing)
	http.HandleFunc("/balance", handler.handleBalance)
	http.HandleFunc("/tx", handler.handleTx)
	http.HandleFunc("/txConfirms", handler.handleTxConfirms)

	portStr := fmt.Sprintf(":%d", util.Constants.HttpPort)
	http.ListenAndServe(portStr, nil)
}
