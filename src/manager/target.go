package manager

import (
	"github.com/levilutz/basiccoin/src/db"
	"github.com/levilutz/basiccoin/src/util"
)

// Create a new mining target block given where to send the reward.
func CreateMiningTarget(s *db.State, inv *db.Inv, publicKeyHash db.HashT) db.Block {
	difficulty, err := db.StringToHash(
		"00000fffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
	)
	if err != nil {
		panic(err)
	}
	// Build tx list until we hit max size
	outTxs := make([]db.Tx, 1)
	outTxs[0] = db.Tx{} // Placeholder for coinbase
	totalFees := uint64(0)
	sizeLeft := util.Constants.MaxBlockVSize - db.CoinbaseVSize()
	candidateTxIds := s.GetSortedIncludableMempool()
	for _, txId := range candidateTxIds {
		tx, ok := inv.LoadTx(txId)
		if !ok {
			panic(db.ErrEntityUnknown)
		}
		// Check if tx is too big to fit in space left
		vSize := tx.VSize()
		if vSize > sizeLeft {
			continue
		}
		// Include tx in out set
		outTxs = append(outTxs, tx)
		sizeLeft -= vSize
		totalFees += tx.TotalInputs() - tx.TotalOutputs()
		// If we're out of space, break
		if sizeLeft < db.MinNonCoinbaseVSize() {
			break
		}
	}
	// Actually make coinbase tx
	headHeight, err := inv.GetBlockHeight(s.GetHead())
	if err != nil {
		panic(err)
	}
	outTxs[0] = db.Tx{
		MinBlock: headHeight + 1,
		Inputs:   make([]db.TxIn, 0),
		Outputs: []db.TxOut{
			{
				Value:         uint64(totalFees) + util.Constants.BlockReward,
				PublicKeyHash: publicKeyHash,
			},
		},
	}
	// Build merkle tree from tx list
	txIds := make([]db.HashT, len(outTxs))
	for i := range txIds {
		txIds[i] = outTxs[i].Hash()
	}
	// Store the coinbase tx
	coinbaseId := outTxs[0].Hash()
	_, ok := inv.LoadTx(coinbaseId)
	if !ok {
		err := inv.StoreTx(outTxs[0])
		if err != nil {
			panic(err)
		}
	}
	s.AddMempoolTx(coinbaseId)
	// Store each merkle node
	merkleMap, merkleIds := db.MerkleFromTxIds(txIds)
	for _, nodeId := range merkleIds {
		_, ok := inv.LoadMerkle(nodeId)
		if ok {
			continue
		}
		err := inv.StoreMerkle(merkleMap[nodeId])
		if err != nil {
			panic(err)
		}
	}
	return db.Block{
		PrevBlockId: s.GetHead(),
		MerkleRoot:  merkleIds[len(merkleIds)-1],
		Difficulty:  difficulty,
	}
}
