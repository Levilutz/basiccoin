package rest

import (
	"net/http"
)

func (s *Server) handleAdminPostTerminate(w http.ResponseWriter, r *http.Request) {
	s.busClient.TerminateCommand()
}
