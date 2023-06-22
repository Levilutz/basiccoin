package rest

import (
	"fmt"
	"io"
	"net/http"

	"github.com/levilutz/basiccoin/internal/pubsub"
)

var adminPrefix = "/admin"
var walletPrefix = "/wallet"

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
	s.mountHandlers(false, "/version", map[string]HttpHandler{
		"GET": func(w http.ResponseWriter, r *http.Request) {
			io.WriteString(w, s.params.Version)
		},
	})

	if s.params.EnableAdmin {
		s.mountHandlers(true, adminPrefix+"/terminate", map[string]HttpHandler{
			"POST": s.handleAdminPostTerminate,
		})
	}

	if s.params.EnableWallet {
		s.mountHandlers(false, walletPrefix+"/balance", map[string]HttpHandler{
			"GET": s.handleWalletGetBalance,
		})
	}

	portStr := fmt.Sprintf(":%d", s.params.Port)
	http.ListenAndServe(portStr, nil)
}

// Mount the given handlers to the given endpoint. Handlers provided as map from methods.
func (s *Server) mountHandlers(admin bool, endpoint string, handlers map[string]HttpHandler) {
	pickMethod := func(w http.ResponseWriter, r *http.Request) {
		if admin && s.params.Password != "" {
			givenPw := r.Header.Get("Pw")
			if givenPw != s.params.Password {
				fmt.Println(givenPw)
				write401(w)
				return
			}
		}
		for method, handler := range handlers {
			if method == r.Method {
				handler(w, r)
				return
			}
		}
		write405(w, r)
	}
	http.HandleFunc(endpoint, pickMethod)
}
