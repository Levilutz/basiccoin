package rest

import (
	"encoding/json"
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
	balance := s.busClient.BalanceQuery(pkh)
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
	utxos := s.busClient.UtxosQuery(pkh)
	utxosJson, err := json.Marshal(utxos)
	if err != nil {
		write500(w, err)
		return
	}
	w.Write(utxosJson)
}

func (s *Server) handleWalletPostTx(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		write400(w, err)
		return
	}
	tx := core.Tx{}
	if err = json.Unmarshal(body, &tx); err != nil {
		write422(w, err)
		return
	}
	if !tx.HasSurplus() {
		write400(w, fmt.Errorf("tx without surplus would never be included"))
		return
	}
	if err = s.busClient.NewTxEvent(tx); err != nil {
		write400(w, err)
		return
	}
	io.WriteString(w, tx.Hash().String())
}
