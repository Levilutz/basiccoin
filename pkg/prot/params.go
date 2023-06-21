package prot

// Params to configure a connection to a peer.
type Params struct {
	Debug          bool   `json:"debug"`
	RuntimeID      string `json:"runtimeId"`      // An id to uniquely identify this node.
	WeAreInitiator bool   `json:"weAreInitiator"` // Whether this peer initiated the connection.
}

// Generate params from the given arguments.
func NewParams(runtimeId string, weAreInitiator bool, debug bool) Params {
	return Params{
		Debug:          debug,
		RuntimeID:      runtimeId,
		WeAreInitiator: weAreInitiator,
	}
}
