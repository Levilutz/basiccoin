package peer

import "github.com/levilutz/basiccoin/internal/pubsub"

func (p *Peer) handleReadAddrsRequest() error {
	p.pubSub.PeersRequested.Pub(pubsub.PeersRequestedEvent{
		PeerRuntimeId: p.conn.PeerRuntimeId(),
	})
	return nil
}

func (p *Peer) handleWriteAddrsRequest() error {
	return nil // cmd was enough
}
