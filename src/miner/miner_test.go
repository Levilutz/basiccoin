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
	difficulty, err := db.StringToHash(
		"00000fffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
	)
	util.AssertNoErr(t, err)
	target := db.Block{
		PrevBlockId: db.HashTZero,
		MerkleRoot:  db.HashTZero,
		Difficulty:  difficulty,
		Nonce:       0,
	}
	solCh := make(chan db.Block)
	miner := NewMiner(target, solCh)
	go miner.Loop()
	timer := time.NewTimer(time.Second * 10)
	select {
	case sol := <-solCh:
		fmt.Println(sol.Nonce)
		fmt.Printf("%x\n", sol.Hash())
		util.Assert(t, db.BelowTarget(sol.Hash(), sol.Difficulty), "not below target")

	case <-timer.C:
		util.Assert(t, false, "timed out mining")
	}
}
