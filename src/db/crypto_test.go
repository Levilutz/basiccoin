package db_test

import (
	"testing"

	. "github.com/levilutz/basiccoin/src/db"
	"github.com/levilutz/basiccoin/src/util"
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
	sigAsn, err := EcdsaSign(priv, NewDHash(content))
	util.AssertNoErr(t, err)
	valid, err := EcdsaVerify(pubDer, NewDHash(content), sigAsn)
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
	sigAsn, err := EcdsaSign(priv, NewDHash(content))
	util.AssertNoErr(t, err)
	valid, err := EcdsaVerify(pubDer, NewDHash(content2), sigAsn)
	util.AssertNoErr(t, err)
	util.Assert(t, !valid, "incorrectly valid signature")
}
