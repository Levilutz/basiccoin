package rest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/levilutz/basiccoin/internal/rest/models"
	"github.com/levilutz/basiccoin/pkg/core"
)

func (s *Server) handleWalletGetBalance(w http.ResponseWriter, r *http.Request) {
	pkhStrs, ok := r.URL.Query()["publicKeyHash"]
	if !ok {
		write400(w, fmt.Errorf("no public key hashes provided"))
		return
	}
	pkhs := make([]core.HashT, len(pkhStrs))
	var err error
	for i, pkhStr := range pkhStrs {
		pkhs[i], err = core.NewHashTFromString(pkhStr)
		if err != nil {
			write400(w, err)
			return
		}
	}
	balances := s.busClient.BalanceQuery(pkhs)
	outJson, err := json.Marshal(models.BalanceResp{
		Balances: balances,
	})
	if err != nil {
		write500(w, err)
		return
	}
	w.Write(outJson)
}

func (s *Server) handleWalletGetUtxos(w http.ResponseWriter, r *http.Request) {
	pkhStrs, ok := r.URL.Query()["publicKeyHash"]
	if !ok {
		write400(w, fmt.Errorf("no public key hashes provided"))
		return
	}
	pkhs := make([]core.HashT, len(pkhStrs))
	var err error
	for i, pkhStr := range pkhStrs {
		pkhs[i], err = core.NewHashTFromString(pkhStr)
		if err != nil {
			write400(w, err)
			return
		}
	}
	utxos := s.busClient.UtxosQuery(pkhs)
	outJson, err := json.Marshal(models.UtxosResp{
		Utxos: utxos,
	})
	if err != nil {
		write500(w, err)
		return
	}
	w.Write(outJson)
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
