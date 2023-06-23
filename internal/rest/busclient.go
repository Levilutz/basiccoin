package rest

import (
	"github.com/levilutz/basiccoin/internal/bus"
	"github.com/levilutz/basiccoin/pkg/core"
)

type BusClient struct {
	bus *bus.Bus
}

func NewBusClient(msgBus *bus.Bus) *BusClient {
	return &BusClient{
		bus: msgBus,
	}
}

func (c *BusClient) NewTxEvent(tx core.Tx) error {
	ret := make(chan error)
	c.bus.CandidateTx.Pub(bus.CandidateTxEvent{
		Ret: ret,
		Tx:  tx,
	})
	return <-ret
}

func (c *BusClient) TerminateCommand() {
	c.bus.Terminate.Pub(bus.TerminateCommand{})
}

func (c *BusClient) BalanceQuery(publicKeyHashes []core.HashT) map[core.HashT]uint64 {
	ret := make(chan map[core.HashT]uint64)
	c.bus.PkhBalance.Pub(bus.PkhBalanceQuery{
		Ret:             ret,
		PublicKeyHashes: publicKeyHashes,
	})
	return <-ret
}

func (c *BusClient) UtxosQuery(publicKeyHashes []core.HashT, excludeMempool bool) map[core.Utxo]core.HashT {
	ret := make(chan map[core.Utxo]core.HashT)
	c.bus.PkhUtxos.Pub(bus.PkhUtxosQuery{
		Ret:             ret,
		PublicKeyHashes: publicKeyHashes,
		ExcludeMempool:  excludeMempool,
	})
	return <-ret
}

func (c *BusClient) TxConfirmsQuery(txIds []core.HashT) map[core.HashT]uint64 {
	ret := make(chan map[core.HashT]uint64)
	c.bus.TxConfirms.Pub(bus.TxConfirmsQuery{
		Ret:   ret,
		TxIds: txIds,
	})
	return <-ret
}
