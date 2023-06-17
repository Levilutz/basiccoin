package miner_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/levilutz/basiccoin/src/db"
	. "github.com/levilutz/basiccoin/src/miner"
	"github.com/levilutz/basiccoin/src/util"
)

func TestMine(t *testing.T) {
	difficulty, err := db.NewHashTFromString(
		"00000fffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
	)
	util.AssertNoErr(t, err)
	merkle := db.NewHashTRand()
	target := db.Block{
		PrevBlockId: db.HashT{},
		MerkleRoot:  merkle,
		Difficulty:  difficulty,
		Nonce:       0,
	}
	m := NewMiner()
	m.SetTarget(target)
	go m.Loop()
	timer := time.NewTimer(time.Second * 10)
	select {
	case sol := <-m.SolutionCh:
		fmt.Println(sol.Nonce)
		fmt.Println(sol.Hash())
		util.Assert(t, sol.Hash().Lt(sol.Difficulty), "not below target")

	case <-timer.C:
		util.Assert(t, false, "timed out mining")
	}
}
