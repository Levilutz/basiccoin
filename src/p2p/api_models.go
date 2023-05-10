package p2p

type VersionResp struct {
	Version     string `json:"version"`
	CurrentTime int64  `json:"currentTime"`
	RuntimeID   string `json:"runtimeId"`
}

type HelloReq struct {
	Addr string `json:"addr"`
}

type AddrIdPair struct {
	Addr      string `json:"addr"`
	RuntimeID string `json:"runtimeID"`
}

type PeersResp struct {
	Peers []AddrIdPair `json:"peers"`
}
