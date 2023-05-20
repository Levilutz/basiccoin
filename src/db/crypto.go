package db

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"fmt"
)

type Hash = [32]byte

// Generate a new hash from the given data.
func NewHash(content ...[]byte) Hash {
	text := make([]byte, 0)
	for _, data := range content {
		text = append(text, data...)
	}
	return sha256.Sum256(text)
}

// Generate a new double hash from the given data.
func NewDHash(content ...[]byte) Hash {
	// Can't one-line bc [:] needs addressable memory
	first := NewHash(content...)
	return NewHash(first[:])
}

func HashHex(hash Hash) string {
	return fmt.Sprintf("%x", hash)
}

// Generate a new ecdsa private key.
func NewEcdsa() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
}

// Marshall an ecdsa private key to base64 of SEC1 ASN.1 DER form.
func MarshallEcdsa(priv *ecdsa.PrivateKey) ([]byte, error) {
	der, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return nil, err
	}
	return EncodeB64(der), nil
}

// Parse an ecdsa private key from base64 of SEC1 ASN.1 DER form.
func ParseECDSA(priv64 []byte) (*ecdsa.PrivateKey, error) {
	der, err := ParseB64(priv64)
	if err != nil {
		return nil, err
	}
	return x509.ParseECPrivateKey(der)
}

// Marshall an ecdsa key's public part to PKIX, ASN.1 DER form.
func MarshallEcdsaPublic(priv *ecdsa.PrivateKey) ([]byte, error) {
	der, err := x509.MarshalPKIXPublicKey(priv.Public())
	if err != nil {
		return nil, err
	}
	return EncodeB64(der), nil
}

// Sign data with ECDSA, return base64-encoded ASN.1 form signature.
// priv is an ecdsa private key.
// hash is the hash of the content that needs to be signed.
func EcdsaSign(priv *ecdsa.PrivateKey, hash Hash) ([]byte, error) {
	sig, err := ecdsa.SignASN1(rand.Reader, priv, hash[:])
	if err != nil {
		return nil, err
	}
	return EncodeB64(sig), nil
}

// Verify an ECDSA signature.
// pub64 is the base64 encoding of PKIX, ASN.1 DER form ecdsa public key.
// hash is the hash of the content that should have been signed.
// sig64 is the base64 encoding of ASN.1 form ecdsa signature.
func EcdsaVerify(pub64 []byte, hash Hash, sig64 []byte) (bool, error) {
	// Parse base64 inputs
	pubRaw, err := ParseB64(pub64)
	if err != nil {
		return false, fmt.Errorf("failed to parse b64 public key: %s", err.Error())
	}
	sig, err := ParseB64(sig64)
	if err != nil {
		return false, fmt.Errorf("failed to parse b64 private key: %s", err.Error())
	}

	// Retrieve public key from DER form
	pubRawKey, err := x509.ParsePKIXPublicKey(pubRaw)
	if err != nil {
		return false, fmt.Errorf("failed to parse DER public key: %s", err.Error())
	}
	pub, ok := pubRawKey.(*ecdsa.PublicKey)
	if !ok {
		return false, fmt.Errorf("unsupported public key type: %T", pubRawKey)
	}

	// Check signature
	return ecdsa.VerifyASN1(pub, hash[:], sig), nil
}

// Encode the given content into base64.
func EncodeB64(content []byte) []byte {
	out := make([]byte, base64.StdEncoding.EncodedLen(len(content)))
	base64.StdEncoding.Encode(out, content)
	return out
}

// Decode content from the given base64, return err if invalid base64.
func ParseB64(content64 []byte) ([]byte, error) {
	out := make([]byte, base64.StdEncoding.DecodedLen(len(content64)))
	n, err := base64.StdEncoding.Decode(out, content64)
	if err != nil {
		return out, err
	}
	return out[:n], nil
}
