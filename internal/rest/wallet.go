package rest

import (
	"io"
	"net/http"
	"strconv"

	"github.com/levilutz/basiccoin/internal/pubsub"
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
	ret := make(chan uint64)
	s.pubSub.PkhBalance.Pub(pubsub.PkhBalanceQuery{
		Ret:           ret,
		PublicKeyHash: pkh,
	})
	balance := <-ret
	io.WriteString(w, strconv.FormatUint(balance, 10))
}
