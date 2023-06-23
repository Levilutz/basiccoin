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

func getVersion(dev bool) string {
	if dev {
		return "v0.0.0-dev"
	} else {
		return "v0.0.0"
	}
}

func NewParams(
	port int, enableAdmin bool, enableWallet bool, password string, dev bool,
) Params {
	return Params{
		Port:         port,
		EnableAdmin:  enableAdmin,
		EnableWallet: enableWallet,
		Password:     password,
		Version:      getVersion(dev),
	}
}
