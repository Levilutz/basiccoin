package rest

import (
	"github.com/levilutz/basiccoin/internal/bus"
	"github.com/levilutz/basiccoin/pkg/core"
)

type PSClient struct {
	bus *bus.Bus
}

func NewPSClient(msgBus *bus.Bus) *PSClient {
	return &PSClient{
		bus: msgBus,
	}
}

func (c *PSClient) BalanceQuery(publicKeyHash core.HashT) uint64 {
	ret := make(chan uint64)
	c.bus.PkhBalance.Pub(bus.PkhBalanceQuery{
		Ret:           ret,
		PublicKeyHash: publicKeyHash,
	})
	return <-ret
}

func (c *PSClient) UtxosQuery(publicKeyHash core.HashT) []core.Utxo {
	ret := make(chan []core.Utxo)
	c.bus.PkhUtxos.Pub(bus.PkhUtxosQuery{
		Ret:           ret,
		PublicKeyHash: publicKeyHash,
	})
	return <-ret
}

func (c *PSClient) TerminateCommand() {
	c.bus.Terminate.Pub(bus.TerminateCommand{})
}
