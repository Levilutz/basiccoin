package chain

import (
	"github.com/levilutz/basiccoin/internal/bus"
	"github.com/levilutz/basiccoin/pkg/core"
	"github.com/levilutz/basiccoin/pkg/set"
)

// Create a new mining target and broadcast it.
func (c *Chain) CreateMiningTarget() {
	// Get candidate txs
	candidateTxIds := c.state.GetSortedIncludableMempool()
	// Build a tx list until we hit max size
	totalSize := core.CoinbaseVSize()
	consumedUtxos := set.NewSet[core.Utxo]()
	txIds := make([]core.HashT, 0)
	for _, txId := range candidateTxIds {
		tx := c.inv.GetTx(txId)
		// Check if tx is too big to fit in remaining space
		vSize := tx.VSize()
		if totalSize+vSize > c.inv.GetCoreParams().MaxBlockVSize {
			continue
		}
		// Check if tx uses already-consumed utxos
		txUtxos := tx.GetConsumedUtxos()
		if consumedUtxos.IncludesAny(txUtxos...) {
			continue
		}
		// Include tx in out set
		txIds = append(txIds, txId)
		totalSize += vSize
		consumedUtxos.Add(txUtxos...)
		// If we couldn't possibly store more txs, stop searching
		if totalSize > c.inv.GetCoreParams().MaxBlockVSize-core.MinNonCoinbaseVSize() {
			break
		}
	}
	c.bus.MinerTarget.Pub(bus.MinerTargetEvent{
		Head:   c.state.head,
		Target: core.NextTarget(c.inv.GetCoreParams(), c.inv, c.state.head),
		TxIds:  txIds,
	})
}
