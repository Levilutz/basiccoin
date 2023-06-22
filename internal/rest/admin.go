package rest

import (
	"net/http"

	"github.com/levilutz/basiccoin/internal/pubsub"
)

func (s *Server) handleAdminPostTerminate(w http.ResponseWriter, r *http.Request) {
	s.pubSub.Terminate.Pub(pubsub.TerminateCommand{})
}
