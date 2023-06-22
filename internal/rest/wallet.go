package rest

import (
	"io"
	"net/http"
	"strconv"
)

func (s *Server) handleWalletGetBalance(w http.ResponseWriter, r *http.Request) {
	balance := uint64(10)
	io.WriteString(w, strconv.FormatUint(balance, 10))
}
