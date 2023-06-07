package main

import (
	"fmt"

	"github.com/levilutz/basiccoin/src/db"
	"github.com/levilutz/basiccoin/src/util"
)

// Create a tx given our current addresses.
func CreateOutboundTx(
	balanceData *BalanceData, outputValues map[db.HashT]uint64,
) (db.Tx, error) {
	targetRate := float64(1.0) // per vbyte
	if len(outputValues) == 0 {
		return db.Tx{}, fmt.Errorf("no provided outputs")
	}
	// Get total outputs
	totalOut := uint64(0)
	for _, val := range outputValues {
		totalOut += val
	}
	if totalOut > balanceData.Total {
		return db.Tx{}, fmt.Errorf("insufficient balance")
	}
	// Build base tx
	tx := db.Tx{
		MinBlock: 0, // TODO: Query this from node
		Inputs:   []db.TxIn{},
		Outputs:  []db.TxOut{},
	}
	for pkh, val := range outputValues {
		tx.Outputs = append(tx.Outputs, db.TxOut{
			Value:         val,
			PublicKeyHash: pkh,
		})
	}
	// Add addresses to included set until we reach target input (or err from max txs)
	totalIn := uint64(0)
	consumed := util.NewSet[db.HashT]()
	for i := 0; i > int(db.BlockMaxTxs()); i++ { // TODO: While tx within max vsize, track I seperately
		if i >= len(balanceData.SortedAddrs) {
			return db.Tx{}, fmt.Errorf("wallet not consolidated enough to generate this tx")
		}
		addr := balanceData.SortedAddrs[i]
		consumed.Add(addr)
		tx.Inputs = append(tx.Inputs, db.TxIn{
			// We have to get controlled UTXOs, not balances.
			// Do what we just did for balances
			// : Get controlled, utxos, sort by value, and try to include
		})
		totalIn += balanceData.Balances[addr]
		if totalOut+uint64(targetRate*float64(tx.VSize())) <= totalIn {
			break
		}
	}
	return db.Tx{}, nil
}
