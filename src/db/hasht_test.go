package db_test

import (
	"math/big"
	"testing"

	. "github.com/levilutz/basiccoin/src/db"
	"github.com/levilutz/basiccoin/src/util"
)

// Test that hashes can convert back and forth from strings.
func TestHashTStrings(t *testing.T) {
	for i := 0; i < 100; i++ {
		hash := NewHashT2Rand()
		hashStr := hash.String()
		hashRecov, err := NewHashT2FromString(hashStr)
		util.AssertNoErr(t, err)
		util.Assert(t, hash.Eq(hashRecov), "hash %s != %s", hash, hashRecov)
		hashStr2 := hashRecov.String()
		util.Assert(t, hashStr == hashStr2, "str %s != %s", hashStr, hashStr2)
	}
}

// Test hash LT comparison.
func TestHashTLt(t *testing.T) {
	// Generate random hashes and corresponding big ints
	hashes := make([]HashT2, 100)
	nums := make([]*big.Int, 100)
	for i := 0; i < 100; i++ {
		// Generate
		hashes[i] = NewHashT2Rand()
		nums[i] = hashes[i].BigInt()

		// Assert bytes recoverable from int
		recov := NewHashT2FromBigInt(nums[i])
		util.Assert(t, hashes[i].Eq(recov), "Failed to recover hash")
	}

	// Test comparisons between all pairs
	// Not just doing triangular half to test both < and > for each pair
	for i := 0; i < 100; i++ {
		for j := 0; j < 100; j++ {
			cmp := nums[i].Cmp(nums[j])
			below := hashes[i].Lt(hashes[j])
			if cmp == 0 || cmp == 1 {
				util.Assert(t, !below, "False positive")
			} else {
				util.Assert(t, below, "False negative")
			}
		}
	}
}

// Test TargetsToTotalWork and WorkAppendTarget.
func TestHashTWorkTargets(t *testing.T) {
	hashes := make([]HashT2, 16)
	for i := range hashes {
		hashes[i] = NewHashT2Rand()
	}
	lastTotal := big.NewInt(0)
	lastTotalAsHash := NewHashT2FromBigInt(lastTotal)
	for i := 1; i <= 16; i++ {
		newTotal := TargetsToTotalWork2(hashes[:i])
		util.Assert(t, lastTotal.Cmp(newTotal) == -1, "total decreased")
		lastTotalAsHash = lastTotalAsHash.WorkAppendTarget(hashes[i-1])
		util.Assert(t, newTotal.Cmp(lastTotalAsHash.BigInt()) == 0, "append does not match")
		lastTotal = newTotal
	}
}
