package rest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/levilutz/basiccoin/src/kern"
	"github.com/levilutz/basiccoin/src/util"
)

type MainQueryHandler interface {
	SyncGetBalance(publicKeyHash kern.HashT) uint64
	SyncGetUtxos(publicKeyHash kern.HashT) []kern.Utxo
	SyncNewTx(tx kern.Tx) error
	SyncGetConfirms(txId kern.HashT) (uint64, bool)
}

type Handler struct {
	m MainQueryHandler
}

func (h *Handler) handleVersion(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		h.handleGetVersion(w, r)
	} else {
		write405(w, r)
	}
}

func (h *Handler) handleGetVersion(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, util.Constants.Version)
}

func (h *Handler) handleBalance(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		h.handleGetBalance(w, r)
	} else {
		write405(w, r)
	}
}

func (h *Handler) handleGetBalance(w http.ResponseWriter, r *http.Request) {
	publicKeyHashes, ok := r.URL.Query()["publicKeyHash"]
	if !ok || len(publicKeyHashes) != 1 {
		write400(w, r, fmt.Errorf("must provide 1 public key hash"))
		return
	}
	pkh, err := kern.NewHashTFromString(publicKeyHashes[0])
	if err != nil {
		write400(w, r, err)
		return
	}
	balance := h.m.SyncGetBalance(pkh)
	io.WriteString(w, strconv.FormatUint(balance, 10))
}

func (h *Handler) handleUtxos(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		h.handleGetUtxos(w, r)
	} else {
		write405(w, r)
	}
}

func (h *Handler) handleGetUtxos(w http.ResponseWriter, r *http.Request) {
	publicKeyHashes, ok := r.URL.Query()["publicKeyHash"]
	if !ok || len(publicKeyHashes) != 1 {
		write400(w, r, fmt.Errorf("must provide 1 public key hash"))
		return
	}
	pkh, err := kern.NewHashTFromString(publicKeyHashes[0])
	if err != nil {
		write400(w, r, err)
		return
	}
	utxos := h.m.SyncGetUtxos(pkh)
	utxosJson, err := json.Marshal(utxos)
	if err != nil {
		write500(w, r, err)
		return
	}
	io.WriteString(w, string(utxosJson))
}

func (h *Handler) handleTx(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		h.handlePostTx(w, r)
	} else {
		write405(w, r)
	}
}

func (h *Handler) handlePostTx(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		write400(w, r, err)
		return
	}
	tx := kern.Tx{}
	if err = json.Unmarshal(body, &tx); err != nil {
		write422(w, r, err)
		return
	}
	if !tx.HasSurplus() {
		write400(w, r, fmt.Errorf("tx without surplus would never be included"))
		return
	}
	if err = h.m.SyncNewTx(tx); err != nil {
		write400(w, r, err)
		return
	}
	io.WriteString(w, fmt.Sprint(tx.Hash()))
}

func (h *Handler) handleTxConfirms(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		h.handleGetTxConfirms(w, r)
	} else {
		write405(w, r)
	}
}

func (h *Handler) handleGetTxConfirms(w http.ResponseWriter, r *http.Request) {
	txIds, ok := r.URL.Query()["txId"]
	if !ok || len(txIds) != 1 {
		write400(w, r, fmt.Errorf("must provide one transaction id"))
		return
	}
	txId, err := kern.NewHashTFromString(txIds[0])
	if err != nil {
		write400(w, r, err)
		return
	}
	confirms, ok := h.m.SyncGetConfirms(txId)
	if !ok {
		write400(w, r, fmt.Errorf("transaction not known"))
		return
	} else {
		io.WriteString(w, strconv.FormatUint(confirms, 10))
	}
}

func write400(w http.ResponseWriter, r *http.Request, err error) {
	w.WriteHeader(400)
	io.WriteString(w, "client error: "+err.Error())
}

func write422(w http.ResponseWriter, r *http.Request, err error) {
	w.WriteHeader(422)
	io.WriteString(w, "failed to parse: "+err.Error())
}

func write405(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(405)
	io.WriteString(w, "method not allowed: "+r.Method)
}

func write500(w http.ResponseWriter, r *http.Request, err error) {
	w.WriteHeader(500)
	io.WriteString(w, "server error: "+err.Error())
}
