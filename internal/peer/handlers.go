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
	return nil
}

func (p *Peer) handleWritePeerAddrs(event pubsub.SendPeersEvent) error {
	return nil
}
