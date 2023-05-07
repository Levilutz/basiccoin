package p2p

import (
	"fmt"
	"sync"
)

type Peers struct {
	mu    sync.Mutex
	peers map[string]*Peer
}

func NewPeers() *Peers {
	return &Peers{
		peers: make(map[string]*Peer),
	}
}

func (pc *Peers) Upsert(addr string, data *PeerData) {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	if peer, ok := pc.peers[addr]; ok {
		peer.UpdateData(data)
	} else {
		pc.peers[addr] = NewPeer(addr, data)
	}
}

func (pc *Peers) IncrementFailures(addr string) (totalFailures int) {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	if peer, ok := pc.peers[addr]; ok {
		return peer.IncrementFailures()
	} else {
		pc.peers[addr] = NewFailedPeer(addr)
		return 1
	}
}

func (pc *Peers) GetData(addr string) (data PeerData, err error) {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	return pc.peers[addr].GetData()
}

func (pc *Peers) GetAddrs() (addrs []string) {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	addrs = make([]string, len(pc.peers))
	i := 0
	for k := range pc.peers {
		addrs[i] = k
		i++
	}
	return
}

func (pc *Peers) DropPeer(addr string) {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	delete(pc.peers, addr)
}

func (pc *Peers) Print() {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	for addr, peer := range pc.peers {
		data, err := peer.GetData()
		var dataStr string
		if err != nil {
			dataStr = err.Error()
		} else {
			dataStr = fmt.Sprintf("%v", data)
		}
		fmt.Printf(
			"%s\t%d\t%s\n",
			addr,
			peer.GetFailures(),
			dataStr,
		)
	}
}
