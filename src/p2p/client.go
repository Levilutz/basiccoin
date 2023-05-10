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
		return 0, errors.New("no peer: " + addr)
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
	pn.mu.Lock()
	defer pn.mu.Unlock()
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

func (pn *P2pNetwork) GetAddrsIds() (addrsIds []AddrIdPair) {
	pn.mu.Lock()
	defer pn.mu.Unlock()
	addrsIds = make([]AddrIdPair, len(pn.peers))
	i := 0
	for addr, peer := range pn.peers {
		addrsIds[i] = AddrIdPair{
			Addr:      addr,
			RuntimeID: peer.GetData().RuntimeID,
		}
	}
	return
}

func (pn *P2pNetwork) GetPeersCopy() map[string]Peer {
	pn.mu.Lock()
	defer pn.mu.Unlock()
	copiedPeers := make(map[string]Peer, len(pn.peers))
	for addr, peer := range pn.peers {
		copiedPeers[addr] = *peer
	}
	return copiedPeers
}

func (pn *P2pNetwork) DropPeer(addr string) {
	pn.mu.Lock()
	defer pn.mu.Unlock()
	delete(pn.peers, addr)
}

func (pn *P2pNetwork) AddPeer(addr string, shouldHello bool) error {
	if pn.HasPeer(addr) {
		return errors.New("cannot add known peer")
	}
	peer, err := DiscoverNewPeer(addr, shouldHello)
	if err != nil {
		return err
	}
	pn.mu.Lock()
	defer pn.mu.Unlock()
	// Re-check that peer wasn't found
	if _, ok := pn.peers[addr]; ok {
		return errors.New("cannot add known peer")
	}
	pn.peers[addr] = peer
	fmt.Println("added peer: " + addr)
	return nil
}

func (pn *P2pNetwork) RetryAddPeer(addr string, shouldHello bool) (err error) {
	for i := 0; i < utils.Constants.AllowedFailures; i++ {
		err = pn.AddPeer(addr, shouldHello)
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

func (pn *P2pNetwork) GetSecondDegree() []string {
	peersCopy := pn.GetPeersCopy()
	// Trigger each peer to get their peers
	results := make(chan AddrIdPair)
	var wg sync.WaitGroup
	for _, peer := range peersCopy {
		peer := peer
		wg.Add(1)
		go func() {
			defer wg.Done()
			addrs, err := peer.GetTheirPeers()
			if err != nil {
				return
			}
			for _, theirPeer := range addrs {
				results <- theirPeer
			}
		}()
	}
	// Collect results from each peer's goroutine
	candidates := make([]AddrIdPair, 0)
	kill := make(chan bool)
	go func() {
		for {
			select {
			case theirPeer := <-results:
				candidates = append(candidates, theirPeer)
			case <-kill:
				return
			}
		}
	}()
	wg.Wait()
	kill <- true
	// Filter for those addrs we don't already have
	result_set := make(map[string]struct{})
	for _, theirPeer := range candidates {
		if _, ok := peersCopy[theirPeer.Addr]; ok {
			continue // Peer known
		}
		if theirPeer.RuntimeID == utils.Constants.RuntimeID {
			continue // Peer is us
		}
		result_set[theirPeer.Addr] = struct{}{}
	}
	actual_results := make([]string, len(result_set))
	for addr := range result_set {
		actual_results = append(actual_results, addr)
	}
	return actual_results
}

func (pn *P2pNetwork) Expand() {
	fmt.Println("seeking new peers...")
	addrs := pn.GetSecondDegree()
	fmt.Printf("found %d potential new peers %v\n", len(addrs), addrs)
	for _, addr := range addrs {
		go pn.AddPeer(addr, true)
	}
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
			if pn.GetCount() < utils.Constants.DesiredPeers {
				go pn.Expand()
			}
		}
	}
}

func (pn *P2pNetwork) Print() {
	pn.mu.Lock()
	defer pn.mu.Unlock()
	fmt.Printf("peers: %d\n", len(pn.peers))
	for addr, peer := range pn.peers {
		fmt.Printf(
			"| %s\t%d\t%v\n",
			addr,
			peer.GetFailures(),
			peer.GetData(),
		)
	}
}
