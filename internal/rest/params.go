package rest

// Params to configure the http server.
type Params struct {
	// What port to host the http server from.
	Port int

	// Whether to enable the admin endpoints.
	EnableAdmin bool

	// Whether to enable the wallet endpoints.
	EnableWallet bool

	// The password to access the admin endpoints.
	Password string

	// The node's current basiccoin version.
	Version string
}

func NewParams(enableAdmin bool, enableWallet bool, password string) Params {
	return Params{
		Port:         8080,
		EnableAdmin:  enableAdmin,
		EnableWallet: enableWallet,
		Password:     password,
		Version:      "v0.0.0",
	}
}
