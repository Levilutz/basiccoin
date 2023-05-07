package p2p

import (
	"time"

	"github.com/levilutz/basiccoin/src/utils"
)

type PeerData struct {
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

func DiscoverNewPeer(addr string) (peer *Peer, err error) {
	data, err := getPeerData(addr)
	if err != nil {
		return nil, err
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
}

func (p *Peer) IncrementFailures() (totalFailures int) {
	p.connectionFailures++
	p.lastUpdated = time.Now()
	return p.connectionFailures
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
		Version:         resp.Version,
		TimeOffsetMicro: resp.CurrentTime - midTimeMicro,
	}, nil
}
