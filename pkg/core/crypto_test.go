package core_test

import (
	"testing"

	. "github.com/levilutz/basiccoin/pkg/core"
	"github.com/levilutz/basiccoin/pkg/util"
)

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
	sigAsn, err := EcdsaSign(priv, DHashBytes(content))
	util.AssertNoErr(t, err)
	valid, err := EcdsaVerify(pubDer, DHashBytes(content), sigAsn)
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
	sigAsn, err := EcdsaSign(priv, DHashBytes(content))
	util.AssertNoErr(t, err)
	valid, err := EcdsaVerify(pubDer, DHashBytes(content2), sigAsn)
	util.AssertNoErr(t, err)
	util.Assert(t, !valid, "incorrectly valid signature")
}
