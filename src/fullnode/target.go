package main

import (
	"time"

	"github.com/levilutz/basiccoin/src/db"
	"github.com/levilutz/basiccoin/src/kern"
	"github.com/levilutz/basiccoin/src/util"
)

// Create a new mining target block given where to send the reward.
// If publicKeyHash is zero value, it's changed to a random hash (used for testing).
func CreateMiningTarget(
	s *db.State, inv *db.Inv, params kern.Params, publicKeyHash kern.HashT,
) kern.Block {
	var err error
	if publicKeyHash.EqZero() {
		publicKeyHash = kern.NewHashTRand()
	}
	target := kern.NextTarget(params, inv, s.GetHead())
	// Build tx list until we hit max size
	outTxs := make([]kern.Tx, 1)
	outTxs[0] = kern.Tx{} // Placeholder for coinbase
	totalFees := uint64(0)
	sizeLeft := params.MaxBlockVSize - kern.CoinbaseVSize()
	candidateTxIds := s.GetSortedIncludableMempool()
	consumedUtxos := util.NewSet[kern.Utxo]()
	for _, txId := range candidateTxIds {
		tx := inv.GetTx(txId)
		// Check if tx is too big to fit in space left
		vSize := tx.VSize()
		if vSize > sizeLeft || consumedUtxos.IncludesAny(tx.GetConsumedUtxos()...) {
			continue
		}
		// Include tx in out set
		outTxs = append(outTxs, tx)
		sizeLeft -= vSize
		// GetSortedIncludableMempool only includes txs with surplus, so this is safe
		totalFees += tx.InputsValue() - tx.OutputsValue()
		consumedUtxos.Add(tx.GetConsumedUtxos()...)
		// If we're out of space, break
		if sizeLeft < kern.MinNonCoinbaseVSize() {
			break
		}
	}
	// Actually make coinbase tx
	headHeight := inv.GetBlockHeight(s.GetHead())
	if err != nil {
		panic(err)
	}
	outTxs[0] = kern.Tx{
		IsCoinbase: true,
		MinBlock:   headHeight + 1,
		Inputs:     make([]kern.TxIn, 0),
		Outputs: []kern.TxOut{
			{
				Value:         uint64(totalFees) + params.BlockReward,
				PublicKeyHash: publicKeyHash,
			},
		},
	}
	// Build merkle kern.HashTm tx list
	txIds := make([]kern.HashT, len(outTxs))
	for i := range txIds {
		txIds[i] = outTxs[i].Hash()
	}
	// Store the coinbase tx
	coinbaseId := outTxs[0].Hash()
	if !inv.HasTx(coinbaseId) {
		err := inv.StoreTx(outTxs[0])
		if err != nil {
			panic(err)
		}
	}
	s.AddMempoolTx(coinbaseId)
	// Store each merkle node
	merkleMap, merkleIds := kern.MerkleFromTxIds(txIds)
	for _, nodeId := range merkleIds {
		if inv.HasMerkle(nodeId) {
			continue
		}
		err := inv.StoreMerkle(merkleMap[nodeId])
		if err != nil {
			panic(err)
		}
	}
	return kern.Block{
		PrevBlockId: s.GetHead(),
		MerkleRoot:  merkleIds[len(merkleIds)-1],
		Target:      target,
		MinedTime:   uint64(time.Now().Unix()),
	}
}
