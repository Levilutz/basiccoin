package peer

import (
	"github.com/levilutz/basiccoin/internal/pubsub"
	"github.com/levilutz/basiccoin/pkg/core"
)

var announceAddrCmd = "announce-addr"

func (p *Peer) handleReadAnnounceAddr() error {
	addr := p.conn.ReadString()
	if p.conn.HasErr() {
		return p.conn.Err()
	}
	p.pubSub.PeerAnnouncedAddr.Pub(pubsub.PeerAnnouncedAddrEvent{
		PeerRuntimeId: p.conn.PeerRuntimeId(),
		Addr:          addr,
	})
	return nil
}

func (p *Peer) handleWriteAnnounceAddr(addr string) error {
	p.conn.WriteString(addr)
	return p.conn.Err()
}

var addrsRequestCmd = "addrs-request"

func (p *Peer) handleReadAddrsRequest() error {
	p.pubSub.PeersRequested.Pub(pubsub.PeersRequestedEvent{
		PeerRuntimeId: p.conn.PeerRuntimeId(),
	})
	return nil
}

func (p *Peer) handleWriteAddrsRequest() error {
	return nil // cmd was enough
}

var peerAddrsCmd = "peer-addrs"

func (p *Peer) handleReadPeerAddrs() error {
	numPeers := p.conn.ReadUint64()
	peerAddrs := make(map[string]string, numPeers)
	for i := 0; i < int(numPeers); i++ {
		runtimeId := p.conn.ReadString()
		peerAddrs[runtimeId] = p.conn.ReadString()
	}
	if p.conn.HasErr() {
		return p.conn.Err()
	}
	p.pubSub.PeersReceived.Pub(pubsub.PeersReceivedEvent{
		PeerAddrs: peerAddrs,
	})
	return nil
}

func (p *Peer) handleWritePeerAddrs(peerAddrs map[string]string) error {
	// Don't change this func without also changing Conn.CloseIfPossible to match.
	p.conn.WriteUint64(uint64(len(peerAddrs)))
	for runtimeId, addr := range peerAddrs {
		p.conn.WriteString(runtimeId)
		p.conn.WriteString(addr)
	}
	return p.conn.Err()
}

var newTxCmd = "new-tx"

func (p *Peer) handleReadNewTx() error {
	txId := p.conn.ReadHashT()
	if p.conn.HasErr() {
		return p.conn.Err()
	} else if p.inv.HasTx(txId) {
		p.conn.WriteBool(false)
		return p.conn.Err()
	}
	p.conn.WriteBool(true)
	tx := p.conn.ReadTx(txId)
	if p.conn.HasErr() {
		return p.conn.Err()
	}
	p.pubSub.CandidateTx.Pub(pubsub.CandidateTxEvent{
		Tx: tx,
	})
	return nil
}

func (p *Peer) handleWriteNewTx(txId core.HashT) error {
	p.conn.WriteHashT(txId)
	wanted := p.conn.ReadBool()
	if p.conn.HasErr() {
		return p.conn.Err()
	} else if wanted {
		p.conn.WriteTx(p.inv.GetTx(txId))
		return p.conn.Err()
	}
	return nil
}
