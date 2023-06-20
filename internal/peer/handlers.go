package peer

import "github.com/levilutz/basiccoin/internal/pubsub"

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

func (p *Peer) handleWritePeerAddrs(event pubsub.SendPeersEvent) error {
	p.conn.WriteUint64(uint64(len(event.PeerAddrs)))
	for runtimeId, addr := range event.PeerAddrs {
		p.conn.WriteString(runtimeId)
		p.conn.WriteString(addr)
	}
	return p.conn.Err()
}
