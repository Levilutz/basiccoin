package rest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/levilutz/basiccoin/internal/rest/models"
	"github.com/levilutz/basiccoin/pkg/core"
)

func (s *Server) handleWalletGetHeadHeight(w http.ResponseWriter, r *http.Request) {
	headHeight := s.busClient.HeadHeightQuery()
	io.WriteString(w, strconv.FormatUint(headHeight, 10))
}

func (s *Server) handleWalletGetTx(w http.ResponseWriter, r *http.Request) {
	txIdStrs, ok := r.URL.Query()["txId"]
	if !ok {
		write400(w, fmt.Errorf("no tx ids provided"))
		return
	}
	txIds, err := core.UnmarshalHashTSlice(txIdStrs)
	if err != nil {
		write400(w, err)
		return
	}
	out := make(map[core.HashT]core.Tx)
	for _, txId := range txIds {
		if s.inv.HasTx(txId) {
			out[txId] = s.inv.GetTx(txId)
		}
	}
	if len(out) == 0 {
		write400(w, fmt.Errorf("no provided tx ids known"))
		return
	}
	outJson, err := json.Marshal(models.GetTxResp{
		Txs: out,
	})
	if err != nil {
		write500(w, err)
		return
	}
	w.Write(outJson)
}

func (s *Server) handleWalletGetMerkle(w http.ResponseWriter, r *http.Request) {
	merkleIdStrs, ok := r.URL.Query()["merkleId"]
	if !ok {
		write400(w, fmt.Errorf("no merkle ids provided"))
		return
	}
	merkleIds, err := core.UnmarshalHashTSlice(merkleIdStrs)
	if err != nil {
		write400(w, err)
		return
	}
	out := make(map[core.HashT]core.MerkleNode)
	for _, merkleId := range merkleIds {
		if s.inv.HasMerkle(merkleId) {
			out[merkleId] = s.inv.GetMerkle(merkleId)
		}
	}
	if len(out) == 0 {
		write400(w, fmt.Errorf("no provided merkle ids known"))
		return
	}
	outJson, err := json.Marshal(models.GetMerkleResp{
		Merkles: out,
	})
	if err != nil {
		write500(w, err)
		return
	}
	w.Write(outJson)
}

func (s *Server) handleWalletGetBlock(w http.ResponseWriter, r *http.Request) {
	blockIdStrs, ok := r.URL.Query()["blockId"]
	if !ok {
		write400(w, fmt.Errorf("no block ids provided"))
		return
	}
	blockIds, err := core.UnmarshalHashTSlice(blockIdStrs)
	if err != nil {
		write400(w, err)
		return
	}
	out := make(map[core.HashT]core.Block)
	for _, blockId := range blockIds {
		if s.inv.HasBlock(blockId) {
			out[blockId] = s.inv.GetBlock(blockId)
		}
	}
	if len(out) == 0 {
		write400(w, fmt.Errorf("no provided block ids known"))
		return
	}
	outJson, err := json.Marshal(models.GetBlockResp{
		Blocks: out,
	})
	if err != nil {
		write500(w, err)
		return
	}
	w.Write(outJson)
}

func (s *Server) handleWalletGetBalance(w http.ResponseWriter, r *http.Request) {
	pkhStrs, ok := r.URL.Query()["publicKeyHash"]
	if !ok {
		write400(w, fmt.Errorf("no public key hashes provided"))
		return
	}
	pkhs, err := core.UnmarshalHashTSlice(pkhStrs)
	if err != nil {
		write400(w, err)
		return
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
	pkhs, err := core.UnmarshalHashTSlice(pkhStrs)
	if err != nil {
		write400(w, err)
		return
	}
	excludeMempool := false
	if vals, ok := r.URL.Query()["excludeMempool"]; ok &&
		len(vals) > 0 && strings.ToLower(vals[0]) == "true" {
		excludeMempool = true
	}
	utxos := s.busClient.UtxosQuery(pkhs, excludeMempool)
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

func (s *Server) handleWalletGetTxConfirms(w http.ResponseWriter, r *http.Request) {
	txIdStrs, ok := r.URL.Query()["txId"]
	if !ok {
		write400(w, fmt.Errorf("no txIds provided"))
		return
	}
	txIds, err := core.UnmarshalHashTSlice(txIdStrs)
	if err != nil {
		write400(w, err)
		return
	}
	confirms := s.busClient.TxConfirmsQuery(txIds)
	outJson, err := json.Marshal(models.TxConfirmsResp{
		Confirms: confirms,
	})
	if err != nil {
		write500(w, err)
		return
	}
	w.Write(outJson)
}

func (s *Server) handleWalletGetTxIncludedBlock(w http.ResponseWriter, r *http.Request) {
	txIdStrs, ok := r.URL.Query()["txId"]
	if !ok {
		write400(w, fmt.Errorf("no txIds provided"))
		return
	}
	txIds, err := core.UnmarshalHashTSlice(txIdStrs)
	if err != nil {
		write400(w, err)
		return
	}
	includedBlocks := s.busClient.TxIncludedBlockQuery(txIds)
	outJson, err := json.Marshal(models.TxIncludedBlockResp{
		IncludedBlocks: includedBlocks,
	})
	if err != nil {
		write500(w, err)
		return
	}
	w.Write(outJson)
}

func (s *Server) handleWalletGetRichList(w http.ResponseWriter, r *http.Request) {
	maxLenStr, ok := r.URL.Query()["maxLen"]
	var maxLen uint64
	var err error
	if ok {
		maxLen, err = strconv.ParseUint(maxLenStr[0], 10, 64)
		if err != nil {
			write400(w, err)
			return
		}
	} else {
		maxLen = 10
	}
	richList := s.busClient.RichListQuery(maxLen)
	outJson, err := json.Marshal(models.RichListResp{
		RichList: richList,
	})
	if err != nil {
		write500(w, err)
		return
	}
	w.Write(outJson)
}
