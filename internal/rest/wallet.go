package rest

import (
	"encoding/json"
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
	balance := s.psClient.BalanceQuery(pkh)
	io.WriteString(w, strconv.FormatUint(balance, 10))
}

func (s *Server) handleWalletGetUtxos(w http.ResponseWriter, r *http.Request) {
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
	utxos := s.psClient.UtxosQuery(pkh)
	utxosJson, err := json.Marshal(utxos)
	if err != nil {
		write500(w, err)
		return
	}
	w.Write(utxosJson)
}
