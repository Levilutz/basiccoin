package rest

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/levilutz/basiccoin/pkg/core"
)

func (s *Server) handleWalletGetBalance(w http.ResponseWriter, r *http.Request) {
	pkhStr, err := getSingleQueryString(w, r, "publicKeyHash")
	if err != nil {
		write400(w, err)
		return
	}
	pkh, err := core.NewHashTFromString(pkhStr)
	if err != nil {
		write400(w, err)
		return
	}
	fmt.Println(pkh)
	balance := uint64(10)
	io.WriteString(w, strconv.FormatUint(balance, 10))
}
