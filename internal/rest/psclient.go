package rest

import (
	"github.com/levilutz/basiccoin/internal/pubsub"
	"github.com/levilutz/basiccoin/pkg/core"
)

type PSClient struct {
	pubSub *pubsub.PubSub
}

func NewPSClient(pubSub *pubsub.PubSub) *PSClient {
	return &PSClient{
		pubSub: pubSub,
	}
}

func (c *PSClient) BalanceQuery(publicKeyHash core.HashT) uint64 {
	ret := make(chan uint64)
	c.pubSub.PkhBalance.Pub(pubsub.PkhBalanceQuery{
		Ret:           ret,
		PublicKeyHash: publicKeyHash,
	})
	return <-ret
}

func (c *PSClient) UtxosQuery(publicKeyHash core.HashT) []core.Utxo {
	ret := make(chan []core.Utxo)
	c.pubSub.PkhUtxos.Pub(pubsub.PkhUtxosQuery{
		Ret:           ret,
		PublicKeyHash: publicKeyHash,
	})
	return <-ret
}

func (c *PSClient) TerminateCommand() {
	c.pubSub.Terminate.Pub(pubsub.TerminateCommand{})
}
