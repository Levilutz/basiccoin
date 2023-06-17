package db

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/hex"
	"fmt"
)

// Generate a new ecdsa private key.
func NewEcdsa() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
}

// Marshall an ecdsa private key to SEC1 ASN.1 DER form.
func MarshalEcdsaPrivate(priv *ecdsa.PrivateKey) ([]byte, error) {
	return x509.MarshalECPrivateKey(priv)
}

// Parse an ecdsa private key from SEC1 ASN.1 DER form.
func ParseECDSAPrivate(priv []byte) (*ecdsa.PrivateKey, error) {
	return x509.ParseECPrivateKey(priv)
}

// Marshall an ecdsa key's public part to PKIX, ASN.1 DER form.
func MarshalEcdsaPublic(priv *ecdsa.PrivateKey) ([]byte, error) {
	return x509.MarshalPKIXPublicKey(priv.Public())
}

// Sign data with ECDSA, return ASN.1 encoded signature.
// priv is an ecdsa private key.
// hash is the hash of the content that needs to be signed.
func EcdsaSign(priv *ecdsa.PrivateKey, hash HashT) ([]byte, error) {
	return ecdsa.SignASN1(rand.Reader, priv, hash.data[:])
}

// Verify an ECDSA signature.
// pub is the DER encoding of PKIX, ASN.1 form ecdsa public key.
// hash is the hash of the content that should have been signed.
// sig is the ASN.1 encoding of ecdsa signature.
func EcdsaVerify(pub []byte, hash HashT, sig []byte) (bool, error) {
	// Retrieve public key from DER form
	pubRawKey, err := x509.ParsePKIXPublicKey(pub)
	if err != nil {
		return false, fmt.Errorf("failed to parse DER public key: %s", err.Error())
	}
	pubKey, ok := pubRawKey.(*ecdsa.PublicKey)
	if !ok {
		return false, fmt.Errorf("unsupported public key type: %T", pubRawKey)
	}

	// Check signature
	return ecdsa.VerifyASN1(pubKey, hash.data[:], sig), nil
}

// An example der-encoded ecdsa public key of expected length 91 bytes.
func ExamplePubDer() []byte {
	pubDer, err := hex.DecodeString(
		"3059301306072a8648ce3d020106082a8648ce3d030107034200042a74cb8265" +
			"947240a77e61fa899cfe7ad0d3ea8df329acd72b22052f0fe4b37b2f5ddbe8a8" +
			"0bb907483121c08db045276e99795db0390d5bcbd80c7bfda68e86",
	)
	if err != nil {
		panic(err)
	}
	return pubDer
}

// An example of a max-length asn.1 encoded ecdsa signature of length 72 bytes.
func ExampleMaxSigAsn() []byte {
	sig, err := hex.DecodeString(
		"3046022100e98da6096a0602d10b6718ff1ce05e396654d427ac5e195c8dfa16" +
			"b43776576b022100d9227403a5fbd8e2ef67b5f22172f94e0b7047d3b5be07e0" +
			"1c671d5d9dfe5a0b",
	)
	if err != nil {
		panic(err)
	}
	return sig
}
