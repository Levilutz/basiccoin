package p2p

import (
	"errors"
	"fmt"
	"time"

	"github.com/levilutz/basiccoin/src/utils"
)

type PeerData struct {
	RuntimeID       string
	Version         string
	TimeOffsetMicro int64
}

type Peer struct {
	addr                    string
	lastUpdated             time.Time
	lastSuccessfullyUpdated time.Time
	connectionFailures      int
	data                    PeerData
}

func NewPeer(addr string, data PeerData) *Peer {
	cTime := time.Now()
	return &Peer{
		addr:                    addr,
		lastUpdated:             cTime,
		lastSuccessfullyUpdated: cTime,
		connectionFailures:      0,
		data:                    data,
	}
}

func DiscoverNewPeer(addr string, shouldHello bool) (peer *Peer, err error) {
	data, err := getPeerData(addr)
	if err != nil {
		return nil, err
	}
	if data.RuntimeID == utils.Constants.RuntimeID ||
		addr == utils.Constants.LocalAddr {
		return nil, errors.New("cannot add self as peer")
	}
	if shouldHello {
		err = utils.PostBody(
			"http://"+addr+"/hello",
			HelloReq{Addr: utils.Constants.LocalAddr},
		)
		if err != nil {
			fmt.Printf("Failed to hello %s: %s\n", addr, err.Error())
		}
	}
	return NewPeer(addr, data), nil
}

func (p *Peer) GetData() (data PeerData) {
	return p.data
}

func (p *Peer) GetFailures() (totalFailures int) {
	return p.connectionFailures
}

func (p *Peer) UpdateData(data PeerData) {
	p.data = data
	cTime := time.Now()
	p.lastUpdated = cTime
	p.lastSuccessfullyUpdated = cTime
	p.connectionFailures = 0
}

func (p *Peer) IncrementFailures() (totalFailures int) {
	p.connectionFailures++
	p.lastUpdated = time.Now()
	return p.connectionFailures
}

func (p *Peer) GetTheirPeers() (addrs []string, err error) {
	resp, _, err := utils.RetryGetBody[PeersResp]("http://"+p.addr+"/peers", 3)
	if err != nil {
		return nil, err
	}
	return resp.Addrs, nil
}

func (p *Peer) Sync() (err error) {
	data, err := getPeerData(p.addr)
	if err != nil {
		p.IncrementFailures()
		return err
	}
	p.UpdateData(data)
	return nil
}

func getPeerData(addr string) (data PeerData, err error) {
	resp, midTimeMicro, err := utils.RetryGetBody[VersionResp](
		"http://"+addr+"/version", 3,
	)
	if err != nil {
		return PeerData{}, err
	}
	return PeerData{
		RuntimeID:       resp.RuntimeID,
		Version:         resp.Version,
		TimeOffsetMicro: resp.CurrentTime - midTimeMicro,
	}, nil
}
