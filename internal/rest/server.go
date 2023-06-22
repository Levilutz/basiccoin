package rest

import (
	"fmt"
	"io"
	"net/http"

	"github.com/levilutz/basiccoin/internal/bus"
)

var adminPrefix = "/admin"
var walletPrefix = "/wallet"

type HttpHandler = func(http.ResponseWriter, *http.Request)

type Server struct {
	params    Params
	busClient *BusClient
}

func NewServer(params Params, msgBus *bus.Bus) *Server {
	return &Server{
		params:    params,
		busClient: NewBusClient(msgBus),
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

		s.mountHandlers(false, walletPrefix+"/utxos", map[string]HttpHandler{
			"GET": s.handleWalletGetUtxos,
		})

		s.mountHandlers(false, walletPrefix+"/tx", map[string]HttpHandler{
			"POST": s.handleWalletPostTx,
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
