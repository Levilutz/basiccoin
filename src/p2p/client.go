package p2p

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/levilutz/basiccoin/src/utils"
)

type P2pNetwork struct {
	mu    sync.Mutex
	peers map[string]*Peer
}

func NewP2pNetwork() *P2pNetwork {
	return &P2pNetwork{
		peers: make(map[string]*Peer),
	}
}

func (pn *P2pNetwork) Upsert(addr string, data PeerData) {
	pn.mu.Lock()
	defer pn.mu.Unlock()
	if peer, ok := pn.peers[addr]; ok {
		peer.UpdateData(data)
	} else {
		pn.peers[addr] = NewPeer(addr, data)
	}
}

func (pn *P2pNetwork) IncrementFailures(addr string) (totalFailures int, err error) {
	pn.mu.Lock()
	defer pn.mu.Unlock()
	if peer, ok := pn.peers[addr]; ok {
		return peer.IncrementFailures(), nil
	} else {
		return 0, errors.New("No peer: " + addr)
	}
}

func (pn *P2pNetwork) GetData(addr string) (data PeerData) {
	pn.mu.Lock()
	defer pn.mu.Unlock()
	return pn.peers[addr].GetData()
}

func (pn *P2pNetwork) HasPeer(addr string) bool {
	pn.mu.Lock()
	defer pn.mu.Unlock()
	_, ok := pn.peers[addr]
	return ok
}

func (pn *P2pNetwork) GetCount() int {
	return len(pn.peers)
}

func (pn *P2pNetwork) GetAddrs() (addrs []string) {
	pn.mu.Lock()
	defer pn.mu.Unlock()
	addrs = make([]string, len(pn.peers))
	i := 0
	for k := range pn.peers {
		addrs[i] = k
		i++
	}
	return
}

func (pn *P2pNetwork) DropPeer(addr string) {
	pn.mu.Lock()
	defer pn.mu.Unlock()
	delete(pn.peers, addr)
}

func (pn *P2pNetwork) AddPeer(addr string) error {
	peer, err := DiscoverNewPeer(addr)
	if err != nil {
		return err
	}
	pn.mu.Lock()
	defer pn.mu.Unlock()
	pn.peers[addr] = peer
	return nil
}

func (pn *P2pNetwork) RetryAddPeer(addr string) (err error) {
	for i := 0; i < utils.Constants.AllowedFailures; i++ {
		err = pn.AddPeer(addr)
		if err == nil {
			return
		}
		time.Sleep(utils.Constants.InitialConnectRetryDelay)
	}
	return err
}

func (pn *P2pNetwork) Sync() {
	pn.mu.Lock()
	defer pn.mu.Unlock()
	var wg sync.WaitGroup
	for addr, peer := range pn.peers {
		addr := addr
		peer := peer
		wg.Add(1)
		go func() {
			defer wg.Done()
			peer.Sync()
			if peer.GetFailures() > utils.Constants.AllowedFailures {
				delete(pn.peers, addr)
			}
		}()
	}
	wg.Wait()
}

func (pn *P2pNetwork) SyncLoop(print bool, kill <-chan bool) {
	ticker := time.NewTicker(utils.Constants.PollingPeriod)
	for {
		select {
		case <-kill:
			return
		case <-ticker.C:
			pn.Sync()
			if print {
				pn.Print()
			}
		}
	}
}

func (pn *P2pNetwork) Print() {
	pn.mu.Lock()
	defer pn.mu.Unlock()
	fmt.Printf("Peers: %d\n", len(pn.peers))
	for addr, peer := range pn.peers {
		fmt.Printf(
			"%s\t%d\t%v\n",
			addr,
			peer.GetFailures(),
			peer.GetData(),
		)
	}
}
