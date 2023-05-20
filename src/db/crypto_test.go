package db_test

import (
	"math/big"
	"testing"

	. "github.com/levilutz/basiccoin/src/db"
	"github.com/levilutz/basiccoin/src/util"
)

// Test that DHashItems and DHashList agree
func TestHashAlternates(t *testing.T) {
	txs := []Tx{
		{
			MinBlock: 5,
		},
		{
			MinBlock: 77,
		},
		{
			MinBlock: 21720,
		},
	}
	txHashes := make([]HashT, len(txs))
	for i, tx := range txs {
		txHashes[i] = tx.Hash()
	}
	baseline := DHashHashes(txHashes)
	allHashes := []HashT{
		DHashItems(txs[0], txs[1], txs[2]),
		DHashItems(txHashes[0], txHashes[1], txHashes[2]),
		DHashList(txs),
	}
	t.Log("baseline", baseline)
	for i, hash := range allHashes {
		util.Assert(
			t, hash == baseline,
			"allHashes[%d] = %s != %s = baseline", i, HashHex(hash), HashHex(baseline),
		)
	}
}

// Test hash hex comparison
func TestBelowTarget(t *testing.T) {
	var err error
	// Generate random hashes and corresponding big ints
	hashes := make([]HashT, 100)
	nums := make([]big.Int, 100)
	for i := 0; i < 100; i++ {
		// Generate
		hashes[i], err = RandHash()
		util.AssertNoErr(t, err)
		nums[i].SetBytes(hashes[i][:])

		// Assert bytes recoverable from int
		recov := make([]byte, 32)
		nums[i].FillBytes(recov)
		recovA := HashT{}
		copy(recovA[:], recov)
		util.Assert(t, hashes[i] == recovA, "Failed to recover hash")
	}

	// Test comparisons between all pairs
	// Not just doing triangular half to test both < and > for each pair
	for i := 0; i < 100; i++ {
		for j := 0; j < 100; j++ {
			cmp := nums[i].Cmp(&nums[j])
			below := BelowTarget(hashes[i], hashes[j])
			if cmp == 0 || cmp == 1 {
				util.Assert(t, !below, "False positive")
			} else {
				util.Assert(t, below, "False negative")
			}
		}
	}
}

// Test generating, marshaling, then parsing a ecdsa private key.
func TestEcdsaReconstruct(t *testing.T) {
	priv, err := NewEcdsa()
	util.AssertNoErr(t, err)
	privDer, err := MarshalEcdsaPrivate(priv)
	util.AssertNoErr(t, err)
	privRecon, err := ParseECDSAPrivate(privDer)
	util.AssertNoErr(t, err)
	util.Assert(
		t, privRecon.Equal(priv) && priv.Equal(privRecon), "key reconstruction failed",
	)
}

// Test that a ecsda signature shows as valid.
func TestEcdsaSign(t *testing.T) {
	priv, err := NewEcdsa()
	util.AssertNoErr(t, err)
	pubDer, err := MarshalEcdsaPublic(priv)
	util.AssertNoErr(t, err)
	content := []byte("Hello World")
	sigAsn, err := EcdsaSign(priv, DHash(content))
	util.AssertNoErr(t, err)
	valid, err := EcdsaVerify(pubDer, DHash(content), sigAsn)
	util.AssertNoErr(t, err)
	util.Assert(t, valid, "invalid signature")
}

// Test that a bad ecdsa signature shows as invalid.
func TestEcdsaBadSign(t *testing.T) {
	priv, err := NewEcdsa()
	util.AssertNoErr(t, err)
	pubDer, err := MarshalEcdsaPublic(priv)
	util.AssertNoErr(t, err)
	content := []byte("Hello World")
	content2 := []byte("Hello World.")
	sigAsn, err := EcdsaSign(priv, DHash(content))
	util.AssertNoErr(t, err)
	valid, err := EcdsaVerify(pubDer, DHash(content2), sigAsn)
	util.AssertNoErr(t, err)
	util.Assert(t, !valid, "incorrectly valid signature")
}
