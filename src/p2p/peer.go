package p2p

import (
	"errors"
	"time"
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
	data                    *PeerData
}

func NewPeer(addr string, data *PeerData) *Peer {
	cTime := time.Now()
	return &Peer{
		addr:                    addr,
		lastUpdated:             cTime,
		lastSuccessfullyUpdated: cTime,
		connectionFailures:      0,
		data:                    data,
	}
}

func NewFailedPeer(addr string) *Peer {
	return &Peer{
		addr:                    addr,
		lastUpdated:             time.Now(),
		lastSuccessfullyUpdated: time.Time{},
		connectionFailures:      1,
		data:                    nil,
	}
}

func (p *Peer) GetData() (data PeerData, err error) {
	if p.data == nil {
		return PeerData{}, errors.New("Peer has no data")
	}
	return *p.data, nil
}

func (p *Peer) GetFailures() (totalFailures int) {
	return p.connectionFailures
}

func (p *Peer) UpdateData(data *PeerData) {
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
