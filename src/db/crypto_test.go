package db_test

import (
	"testing"

	. "github.com/levilutz/basiccoin/src/db"
)

func assert(t *testing.T, condition bool, msg string, v ...interface{}) {
	if !condition {
		t.Fatalf(msg, v...)
	}
}

func assertNoErr(t *testing.T, err error) {
	errStr := ""
	if err != nil {
		errStr = err.Error()
	}
	assert(t, err == nil, "unexpected error: %s", errStr)
}

// Test generating, marshaling, then parsing a ecdsa private key.
func TestEcdsaReconstruct(t *testing.T) {
	priv, err := NewEcdsa()
	assertNoErr(t, err)
	priv64, err := MarshallEcdsa(priv)
	assertNoErr(t, err)
	privRecon, err := ParseECDSA(priv64)
	assertNoErr(t, err)
	assert(
		t, privRecon.Equal(priv) && priv.Equal(privRecon), "key reconstruction failed",
	)
}

// Test that a ecsda signature shows as valid.
func TestEcdsaSign(t *testing.T) {
	priv, err := NewEcdsa()
	assertNoErr(t, err)
	pub64, err := MarshallEcdsaPublic(priv)
	assertNoErr(t, err)
	content := []byte("Hello World")
	sig, err := EcdsaSign(priv, NewDHash(content))
	assertNoErr(t, err)
	valid, err := EcdsaVerify(pub64, NewDHash(content), sig)
	assertNoErr(t, err)
	assert(t, valid, "invalid signature")
}

// Test that a bad ecdsa signature shows as invalid.
func TestEcdsaBadSign(t *testing.T) {
	priv, err := NewEcdsa()
	assertNoErr(t, err)
	pub64, err := MarshallEcdsaPublic(priv)
	assertNoErr(t, err)
	content := []byte("Hello World")
	content2 := []byte("Hello World.")
	sig, err := EcdsaSign(priv, NewDHash(content))
	assertNoErr(t, err)
	valid, err := EcdsaVerify(pub64, NewDHash(content2), sig)
	assertNoErr(t, err)
	assert(t, !valid, "incorrectly valid signature")
}
