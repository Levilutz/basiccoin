package rest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/levilutz/basiccoin/src/db"
)

type MainQueryHandler interface {
	SyncGetBalance(publicKeyHash db.HashT) uint64
	SyncNewTx(tx db.Tx) error
}

type Handler struct {
	m MainQueryHandler
}

func (h *Handler) handleGetPing(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "pong")
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
	if !ok || len(publicKeyHashes) < 1 {
		write400(w, r, fmt.Errorf("must provide public key hash"))
		return
	}
	pkh, err := db.StringToHash(publicKeyHashes[0])
	if err != nil {
		write400(w, r, err)
		return
	}
	balance := h.m.SyncGetBalance(pkh)
	io.WriteString(w, strconv.FormatUint(balance, 10))
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
	tx := db.Tx{}
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
