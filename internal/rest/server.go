package rest

import (
	"fmt"
	"io"
	"net/http"

	"github.com/levilutz/basiccoin/internal/pubsub"
)

type HttpHandler = func(http.ResponseWriter, *http.Request)

type Server struct {
	params Params
	pubSub *pubsub.PubSub
}

func NewServer(params Params, pubSub *pubsub.PubSub) *Server {
	return &Server{
		params: params,
		pubSub: pubSub,
	}
}

func (s *Server) Start() {
	mountHandlers("/version", map[string]HttpHandler{
		"GET": func(w http.ResponseWriter, r *http.Request) {
			io.WriteString(w, s.params.Version)
		},
	})

	portStr := fmt.Sprintf(":%d", s.params.Port)
	http.ListenAndServe(portStr, nil)
}

// Mount the given handlers to the given endpoint. Handlers provided as map from methods.
func mountHandlers(endpoint string, handlers map[string]HttpHandler) {
	pickMethod := func(w http.ResponseWriter, r *http.Request) {
		for method, handler := range handlers {
			if method == r.Method {
				handler(w, r)
				return
			}
		}
		w.WriteHeader(405)
		io.WriteString(w, "method not allowed: "+r.Method)
	}
	http.HandleFunc(endpoint, pickMethod)
}
