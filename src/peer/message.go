package peer

type PeerMessage interface {
	Transmit(pc *PeerConn) error
}

// HelloMessage

type HelloMessage struct {
	RuntimeID string `json:"runtimeID"`
	Version   string `json:"version"`
	Addr      string `json:"addr"`
}
