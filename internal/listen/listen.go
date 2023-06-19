package listen

import "github.com/levilutz/basiccoin/internal/pubsub"

type Listener struct {
	pubSub *pubsub.PubSub
}

func NewListener(pubSub *pubsub.PubSub) *Listener {
	return &Listener{
		pubSub: pubSub,
	}
}
