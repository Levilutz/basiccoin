package peerfactory

import "github.com/levilutz/basiccoin/internal/pubsub"

type PeerFactory struct {
	pubSub *pubsub.PubSub
}

func NewPeerFactory(pubSub *pubsub.PubSub) *PeerFactory {
	return &PeerFactory{
		pubSub: pubSub,
	}
}
