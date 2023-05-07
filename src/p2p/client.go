package p2p

import (
	"fmt"
	"sync"
	"time"
)

type PeerData struct {
	Version         string
	TimeOffsetMicro int64
}

type PeerRecord struct {
	lastSuccessfullyUpdated time.Time
	connectionFailures      int
	data                    *PeerData
}

type Peers struct {
	mu    sync.Mutex
	peers map[string]PeerRecord
}

func NewPeers() *Peers {
	return &Peers{
		peers: make(map[string]PeerRecord),
	}
}

func (pc *Peers) Upsert(addr string, data *PeerData) {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	pc.peers[addr] = PeerRecord{time.Now(), 0, data}
}

func (pc *Peers) IncrementFailures(addr string) (totalFailures int) {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	if record, ok := pc.peers[addr]; ok {
		record.connectionFailures++
		totalFailures = record.connectionFailures
		pc.peers[addr] = record
	} else {
		pc.peers[addr] = PeerRecord{time.Time{}, 1, nil}
		totalFailures = 1
	}
	return
}

func (pc *Peers) Get(addr string) (record PeerRecord) {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	record = pc.peers[addr]
	return
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
	for addr, record := range pc.peers {
		fmt.Printf(
			"%s\t%d\t%d\t%v\n",
			addr,
			record.connectionFailures,
			record.lastSuccessfullyUpdated.Unix(),
			record.data,
		)
	}
}
