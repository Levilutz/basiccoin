package db

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/levilutz/basiccoin/src/util"
)

type Hasher interface {
	Hash() HashT
}

type HashT = [32]byte

var HashTZero = HashT{}

// Generate a random hash
func RandHash() (HashT, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return HashT{}, err
	}
	out := HashT{}
	copy(out[:], bytes)
	return out, nil
}

// Generate a new hash from the given data.
func singleHash(content ...[]byte) HashT {
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
func DHash(content ...[]byte) HashT {
	// Can't one-line bc [:] needs addressable memory
	first := singleHash(content...)
	return singleHash(first[:])
}

// Generate a new double hash from the given uint64
func DHashUint64(value uint64) HashT {
	bs := make([]byte, 8)
	binary.BigEndian.PutUint64(bs, value)
	return DHash(bs)
}

// Generate a double hash from a list of existing hahes
func DHashHashes(items []HashT) HashT {
	itemHashes := make([][]byte, len(items))
	for i := 0; i < len(items); i++ {
		itemHashes[i] = items[i][:]
	}
	return DHash(itemHashes...)
}

// Hash from a list of hasher inputs
func DHashList[T Hasher](items []T) HashT {
	itemHashes := make([][]byte, len(items))
	for i := 0; i < len(items); i++ {
		itemHash := items[i].Hash()
		itemHashes[i] = itemHash[:]
	}
	return DHash(itemHashes...)
}

// Generate root-node hash of depth-1 tree, given children.
// If child is a hash, it's included unchanged.
// If child is a Hashable, it's Hash() method is run and the output is included.
// If child is an int, it's converted to big endian bytes, hashed, then included.
// If child is a []byte, it's hashed normally.
// If unknown type, it panics (should be unreachable).
func DHashItems(children ...any) HashT {
	itemHashes := make([][]byte, len(children))
	for i := 0; i < len(children); i++ {
		var itemHash HashT
		switch item := children[i].(type) {
		case Hasher:
			itemHash = item.Hash()
		case HashT:
			itemHash = item
		case []byte:
			itemHash = DHash(item)
		case uint64:
			itemHash = DHashUint64(item)
		default:
			panic(fmt.Sprintf("unhashable type: %T", item))
		}
		itemHashes[i] = itemHash[:]
	}
	return DHash(itemHashes...)
}

// Generate hex string representation of hash
func HashHex(hash HashT) string {
	return fmt.Sprintf("%x", hash)
}

// Whether a < b, big-endian.
// Also whether a given hash (a) is below a target (b).
func HashLT(a HashT, b HashT) bool {
	for i := 0; i < 32; i++ {
		if a[i] > b[i] {
			return false
		} else if a[i] < b[i] {
			return true
		}
	}
	// Values equal
	return false
}

// Compute total work from a set of targets.
func TargetsToTotalWork(targets []HashT) *big.Int {
	total := &big.Int{}
	for _, target := range targets {
		if target == HashTZero {
			panic("cannot compute work for zero target")
		}
		// total += 2^256 / target
		targetInt := &big.Int{}
		targetInt.SetBytes(target[:])
		targetInt.Div(util.BigInt2_256(), targetInt)
		total.Add(total, targetInt)
	}
	return total
}

// Compute new total work given the addition of a new target difficulty.
// Returns prior + (2^32 / target).
func AppendTotalWork(prior HashT, target HashT) HashT {
	// Convert to big ints
	priorInt := &big.Int{}
	targetInt := &big.Int{}
	priorInt.SetBytes(prior[:])
	targetInt.SetBytes(target[:])
	// Compute work from new target
	targetInt.Div(util.BigInt2_256(), targetInt)
	priorInt.Add(priorInt, targetInt)
	// Convert back to HashT to return
	priorInt.FillBytes(prior[:])
	return prior
}

func StringToHash(data string) (HashT, error) {
	if len(data) != 64 {
		return HashT{}, fmt.Errorf("cannot parse hash from length %d", len(data))
	}
	out, err := hex.DecodeString(data)
	if err != nil {
		return HashT{}, err
	}
	outP := (*HashT)(out)
	return *outP, nil
}

func StringToHashes(data string, numHashes int) ([]HashT, error) {
	if len(data) != numHashes*64 {
		return nil, fmt.Errorf(
			"expected length %d != actual length %d",
			numHashes*64,
			len(data),
		)
	}
	out := make([]HashT, numHashes)
	for i := 0; i < numHashes; i++ {
		hash, err := StringToHash(data[i*64 : (i+1)*64])
		if err != nil {
			return nil, err
		}
		out[i] = hash
	}
	return out, nil
}

func HashesToString(hashes []HashT) string {
	out := ""
	for i := 0; i < len(hashes); i++ {
		out += fmt.Sprintf("%x", hashes[i])
	}
	return out
}

func HasherMap[K Hasher](list []K) map[HashT]K {
	out := make(map[HashT]K)
	for _, item := range list {
		out[item.Hash()] = item
	}
	return out
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
