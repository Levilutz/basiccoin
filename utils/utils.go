package utils

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"math/big"
)

func Hash(texts ...[]byte) []byte {
	h := sha256.New()
	for _, text := range texts {
		h.Write(text)
	}
	return h.Sum(nil)
}

func Dhash(texts ...[]byte) []byte {
	return Hash(Hash(texts...))
}

func Concat(texts ...[]byte) []byte {
	totLen := 0
	for _, text := range texts {
		totLen += len(text)
	}
	out := make([]byte, totLen)
	i := 0
	for _, text := range texts {
		copy(out[i:], text)
		i += len(text)
	}
	return out
}

func Ecdsa256() *ecdsa.PrivateKey {
	privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	return privateKey
}

func EcdsaToKeys(privateKey *ecdsa.PrivateKey) ([]byte, []byte, []byte) {
	publicKey := privateKey.PublicKey
	return privateKey.D.Bytes(), publicKey.X.Bytes(), publicKey.Y.Bytes()
}

func KeysToEcdsa(privateKey, publicKeyX, publicKeyY []byte) *ecdsa.PrivateKey {
	return &ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{
			Curve: elliptic.P256(),
			X:     new(big.Int).SetBytes(publicKeyX),
			Y:     new(big.Int).SetBytes(publicKeyY),
		},
		D: new(big.Int).SetBytes(privateKey),
	}
}
