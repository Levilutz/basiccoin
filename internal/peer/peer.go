package peer

import "github.com/levilutz/basiccoin/internal/pubsub"

type Peer struct {
	pubSub *pubsub.PubSub
}

func NewPeer(pubSub *pubsub.PubSub) *Peer {
	return &Peer{
		pubSub: pubSub,
	}
}
