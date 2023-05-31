package db_test

import (
	"testing"

	. "github.com/levilutz/basiccoin/src/db"
	"github.com/levilutz/basiccoin/src/util"
)

func TestMerkleFromTxIds(t *testing.T) {
	var err error
	// Build list of txs
	txIds := make([]HashT, 5)
	for i := 0; i < 5; i++ {
		txIds[i], err = RandHash()
		util.AssertNoErr(t, err)
	}
	// Construct merkle tree
	merkleMap, merkleIds := MerkleFromTxIds(txIds)
	util.Assert(t, len(merkleMap) == len(merkleIds), "out lengths mismatched")
	util.Assert(t, len(merkleIds) == 6, "out length unexpected")
	util.Assert(t, len(merkleIds) <= int(MerkleTreeMaxSize()), "tree max size wrong")
	expected := []MerkleNode{
		{
			LChild: txIds[0],
			RChild: txIds[1],
		},
		{
			LChild: txIds[2],
			RChild: txIds[3],
		},
		{
			LChild: txIds[4],
			RChild: txIds[4],
		},
	}
	expected = append(expected, MerkleNode{
		LChild: expected[0].Hash(),
		RChild: expected[1].Hash(),
	})
	expected = append(expected, MerkleNode{
		LChild: expected[2].Hash(),
		RChild: expected[2].Hash(),
	})
	expected = append(expected, MerkleNode{
		LChild: expected[3].Hash(),
		RChild: expected[4].Hash(),
	})
	for i := 0; i < 6; i++ {
		nodeId := expected[i].Hash()
		util.Assert(t, merkleIds[i] == nodeId, "id mismatch")
		util.Assert(t, merkleMap[nodeId] == expected[i], "content mismatch")
	}
}