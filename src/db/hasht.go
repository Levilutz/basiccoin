package db

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
)

// A container for Hash values.
// Was long just a type alias for [32]byte, but giving it methods makes life easier.
type HashT2 struct {
	data [32]byte
}

// Generate a new random hash.
func NewHashT2Rand() HashT2 {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		panic(err)
	}
	out := HashT2{}
	copy(out.data[:], bytes)
	return out
}

// Parse a hash from a hex-encoding in a string.
func NewHashT2FromString(data string) (HashT2, error) {
	if len(data) != 64 {
		return HashT2{}, fmt.Errorf("cannot parse hash from length %d", len(data))
	}
	decoded, err := hex.DecodeString(data)
	if err != nil {
		return HashT2{}, err
	}
	out := HashT2{}
	copy(out.data[:], decoded)
	return out, nil
}

// Create a hash from a big.Int. Panics if data > 2^32-1
func NewHashT2FromBigInt(data *big.Int) HashT2 {
	out := HashT2{}
	data.FillBytes(out.data[:])
	return out
}

// Retrieve the underlying byte array from the HashT.
func (h HashT2) Data() [32]byte {
	return h.data
}

// Convert to a hex-encoded string.
func (h HashT2) String() string {
	return fmt.Sprintf("%x", h.data)
}

// Check whether this hash is equal in value to another.
func (h HashT2) Eq(other HashT2) bool {
	return h.data == other.data
}

// Check whether this hash is less than another (big-endian).
func (h HashT2) Lt(other HashT2) bool {
	for i := 0; i < 32; i++ {
		if h.data[i] > other.data[i] {
			return false
		} else if h.data[i] < other.data[i] {
			return true
		}
	}
	// Values equal
	return false
}

// Convert the Hash to a big.Int.
func (h HashT2) BigInt() *big.Int {
	out := &big.Int{}
	out.SetBytes(h.data[:])
	return out
}

// Convert a hash target to a big.Int amount of work expected to beat it.
func (h HashT2) TargetToWork() *big.Int {
	if h.data == [32]byte{} {
		panic("cannot compute work for zero target")
	}
	// return 2^256 / h.data
	out := h.BigInt()
	out.Div(bigInt2_256(), out)
	return out
}

// Increase the given total amount of work by the given target's work.
func (h HashT2) WorkAppendTarget(newTarget HashT2) HashT2 {
	curInt := h.BigInt()
	nextInt := newTarget.TargetToWork()
	curInt.Add(curInt, nextInt)
	return NewHashT2FromBigInt(curInt)
}

// Convert the given list of targets to an amount of total work to reach them all.
func TargetsToTotalWork2(targets []HashT2) *big.Int {
	total := &big.Int{}
	for _, target := range targets {
		total.Add(total, target.TargetToWork())
	}
	return total
}

// Compute 2^256 as a big.Int.
func bigInt2_256() *big.Int {
	out := &big.Int{}
	out.SetString(
		"ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
		16,
	)
	return out
}

// Any object that defines how it is meant to be hashed.
type Hasher2 interface {
	Hash2() HashT2
}

// Generate a new double-sha256 hash from the given bytes.
func DHashBytes2(content []byte) HashT2 {
	// Can't one-line bc [:] needs addressable memory
	first := sha256.Sum256(content)
	return HashT2{
		data: sha256.Sum256(first[:]),
	}
}

// Generate a new double-sha256 hash from the given uint64.
func DHashUint642(content uint64) HashT2 {
	bs := make([]byte, 8)
	binary.BigEndian.PutUint64(bs, content)
	return DHashBytes2(bs)
}

// Generate a new double-sha256 hash from whatever the given value is.
// If content is a hash, it's returned unchanged.
// If content is a Hasher, the output of its Hash() method is returned.
// If content is a uint64, the hash of its big-endian bytes is returned.
// If content is a byte slice, its hash is returned.
// If content is of unexpected type, this method panics.
func DHashAny2(content any) HashT2 {
	switch typed := content.(type) {
	case Hasher2:
		return typed.Hash2()
	case HashT2:
		return typed
	case uint64:
		return DHashUint642(typed)
	case []byte:
		return DHashBytes2(typed)
	default:
		panic(fmt.Sprintf("unhashable type: %T", typed))
	}
}

// Generate a new double-sha256 hash of the given hashes, all concatenated.
func DHashHashes2(items []HashT2) HashT2 {
	concat := make([]byte, len(items)*32)
	for _, item := range items {
		concat = append(concat, item.data[:]...)
	}
	return DHashBytes2(concat)
}

// Generate a new double-sha256 hash of the given various items concatenated.
// See DHashAny2 for documentation on how each type is handled
func DHashVarious2(items ...any) HashT2 {
	hashes := make([]HashT2, len(items))
	for i := range items {
		hashes[i] = DHashAny2(items[i])
	}
	return DHashHashes2(hashes)
}

func DHashList2[T any](items []T) HashT2 {
	hashes := make([]HashT2, len(items))
	for i := range items {
		hashes[i] = DHashAny2(items[i])
	}
	return DHashHashes2(hashes)
}
