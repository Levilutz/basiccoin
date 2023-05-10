package p2p

type VersionResp struct {
	Version     string `json:"version"`
	CurrentTime int64  `json:"currentTime"`
	RuntimeID   string `json:"runtimeId"`
}

type HelloReq struct {
	Addr string `json:"addr"`
}

type PeersResp struct {
	Addrs []string `json:"addrs"`
}
