package db_test

import (
	"fmt"
	"math/big"
	"testing"

	. "github.com/levilutz/basiccoin/src/db"
	"github.com/levilutz/basiccoin/src/util"
)

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
			below := HashLT(hashes[i], hashes[j])
			if cmp == 0 || cmp == 1 {
				util.Assert(t, !below, "False positive")
			} else {
				util.Assert(t, below, "False negative")
			}
		}
	}
}

// Test hash targets to total work
func TestTargetsToTotalWork(t *testing.T) {
	hashes := []HashT{
		{
			205, 193, 55, 231, 234, 91, 89, 206, 201, 24, 32, 213, 16, 237, 38, 176,
			126, 143, 125, 138, 224, 54, 162, 4, 179, 78, 35, 109, 252, 132, 213, 174,
		},
		{
			0, 0, 126, 78, 212, 166, 109, 71, 60, 53, 53, 27, 166, 218, 20, 34,
			183, 228, 60, 18, 12, 75, 168, 201, 88, 4, 135, 229, 246, 32, 52, 214,
		},
		{
			100, 185, 149, 76, 42, 55, 218, 95, 119, 112, 160, 94, 53, 7, 180, 255,
			103, 46, 231, 0, 144, 245, 27, 34, 181, 196, 110, 134, 94, 155, 167, 214,
		},
		{
			99, 22, 213, 54, 151, 89, 64, 78, 247, 142, 14, 250, 176, 92, 111, 66,
			138, 47, 15, 141, 167, 248, 228, 29, 37, 251, 172, 172, 120, 131, 103, 164,
		},
		{
			0, 0, 0, 0, 0, 152, 61, 205, 87, 255, 88, 152, 171, 153, 146, 129,
			28, 51, 100, 132, 25, 248, 240, 118, 59, 162, 205, 186, 32, 150, 150, 255,
		},
		{
			131, 25, 42, 86, 23, 205, 245, 18, 55, 13, 212, 231, 29, 148, 77, 66,
			176, 213, 204, 209, 249, 37, 37, 169, 12, 12, 235, 175, 250, 102, 193, 156,
		},
		{
			58, 65, 220, 144, 174, 104, 20, 140, 54, 43, 188, 130, 194, 252, 87, 169,
			101, 173, 192, 238, 97, 32, 36, 130, 54, 50, 84, 234, 29, 205, 217, 131,
		},
		{
			0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 135, 222, 64, 19, 185, 197,
			81, 9, 66, 95, 66, 78, 189, 57, 32, 58, 19, 215, 127, 126, 136, 228,
		},
		{
			142, 96, 48, 129, 244, 248, 26, 56, 118, 89, 164, 237, 31, 254, 121, 37,
			154, 213, 14, 253, 105, 109, 218, 87, 182, 153, 214, 67, 206, 208, 32, 149,
		},
		{
			38, 179, 184, 74, 169, 43, 82, 29, 28, 218, 76, 62, 53, 236, 224, 80,
			18, 109, 252, 130, 166, 117, 195, 63, 230, 249, 33, 23, 63, 238, 207, 251,
		},
		{
			202, 74, 64, 52, 211, 25, 247, 172, 144, 235, 53, 237, 165, 157, 249, 100,
			95, 115, 149, 57, 74, 111, 189, 230, 23, 253, 224, 164, 184, 21, 68, 176,
		},
		{
			116, 31, 51, 133, 83, 180, 244, 89, 53, 216, 118, 79, 203, 32, 110, 148,
			23, 168, 5, 74, 125, 213, 59, 201, 51, 190, 16, 91, 245, 99, 18, 49,
		},
		{
			6, 83, 15, 190, 185, 225, 55, 230, 176, 89, 187, 153, 55, 194, 206, 209,
			194, 116, 64, 187, 40, 42, 162, 49, 68, 60, 167, 95, 215, 10, 90, 162,
		},
		{
			38, 43, 219, 173, 167, 194, 225, 29, 111, 78, 192, 221, 38, 85, 194, 45,
			193, 155, 12, 167, 189, 55, 233, 173, 1, 225, 25, 125, 15, 92, 190, 137,
		},
		{
			0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			10, 180, 111, 37, 11, 164, 165, 31, 219, 105, 137, 136, 195, 199, 48, 249,
		},
		{
			34, 13, 103, 61, 75, 19, 15, 176, 144, 230, 95, 196, 225, 58, 151, 226,
			204, 122, 252, 249, 11, 183, 13, 14, 235, 107, 161, 43, 34, 203, 41, 230,
		},
	}
	lastTotal := big.NewInt(0)
	for i := 1; i <= 16; i++ {
		newTotal := TargetsToTotalWork(hashes[:i])
		t.Logf("%d %s %s %x\n", i, lastTotal.String(), newTotal.String(), hashes[i-1])
		util.Assert(t, lastTotal.Cmp(newTotal) == -1, "total decreased")
		lastTotal = newTotal
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
	sigAsn, err := EcdsaSign(priv, DHashBytes2(content))
	util.AssertNoErr(t, err)
	valid, err := EcdsaVerify(pubDer, DHashBytes2(content), sigAsn)
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
	sigAsn, err := EcdsaSign(priv, DHashBytes2(content))
	util.AssertNoErr(t, err)
	valid, err := EcdsaVerify(pubDer, DHashBytes2(content2), sigAsn)
	util.AssertNoErr(t, err)
	util.Assert(t, !valid, "incorrectly valid signature")
}

// Test that hashes can be read from strings correctly.
func TestStringToHash(t *testing.T) {
	for i := 0; i < 100; i++ {
		hash, err := RandHash()
		util.AssertNoErr(t, err)
		hashStr := fmt.Sprintf("%x", hash)
		util.Assert(t, len(hashStr) == 64, "hash hex length 64 != %d", len(hashStr))
		hashRecov, err := StringToHash(hashStr)
		util.AssertNoErr(t, err)
		util.Assert(t, hash == hashRecov, "%x != %x", hash, hashRecov)
	}
}

func TestStringToHashes(t *testing.T) {
	hashes := make([]HashT, 100)
	hashesStr := ""
	for i := 0; i < 100; i++ {
		hash, err := RandHash()
		util.AssertNoErr(t, err)
		hashes[i] = hash
		hashStr := fmt.Sprintf("%x", hash)
		util.Assert(t, len(hashStr) == 64, "hash hex length 64 != %d", len(hashStr))
		hashesStr += hashStr

	}
	hashesRecov, err := StringToHashes(hashesStr, 100)
	util.AssertNoErr(t, err)
	for i := 0; i < 100; i++ {
		util.Assert(
			t, hashes[i] == hashesRecov[i], "%x != %x", hashes[i], hashesRecov[i],
		)
	}
}
