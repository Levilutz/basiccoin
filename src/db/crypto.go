package db

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"strconv"
)

type Hasher interface {
	Hash() HashT
}

type HashT = [32]byte

// Generate a new hash from the given data.
func NewHash(content ...[]byte) HashT {
	if len(content) == 1 {
		return sha256.Sum256(content[0])
	}
	text := make([]byte, 0)
	for _, data := range content {
		text = append(text, data...)
	}
	return sha256.Sum256(text)
}

// Generate a new double hash from the given data.
func NewDHash(content ...[]byte) HashT {
	// Can't one-line bc [:] needs addressable memory
	first := NewHash(content...)
	return NewHash(first[:])
}

// Generate a new double hash from the given int (encoded as str)
func NewDHashInt(value int) HashT {
	return NewDHash([]byte(strconv.Itoa(value)))
}

// Hash from a list of hasher inputs
func NewDHashList[T Hasher](items []T) HashT {
	itemHashes := make([][]byte, len(items))
	for i := 0; i < len(items); i++ {
		itemHash := items[i].Hash()
		itemHashes[i] = itemHash[:]
	}
	return NewDHash(itemHashes...)
}

// Generate hex string representation of hash
func HashHex(hash HashT) string {
	return fmt.Sprintf("%x", hash)
}

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
	return ecdsa.SignASN1(rand.Reader, priv, hash[:])
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
	return ecdsa.VerifyASN1(pubKey, hash[:], sig), nil
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
