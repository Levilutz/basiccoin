package core

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
)

// A container for Hash values.
// Was long just a type alias for [32]byte, but giving it methods makes life easier.
type HashT struct {
	data [32]byte
}

// Generate a new random hash.
func NewHashTRand() HashT {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		panic(err)
	}
	out := HashT{}
	copy(out.data[:], bytes)
	return out
}

// Parse a hash from a hex-encoding in a string.
func NewHashTFromString(data string) (HashT, error) {
	if len(data) != 64 {
		return HashT{}, fmt.Errorf("cannot parse hash from length %d", len(data))
	}
	decoded, err := hex.DecodeString(data)
	if err != nil {
		return HashT{}, err
	}
	out := HashT{}
	copy(out.data[:], decoded)
	return out, nil
}

// Parse a hash from a hex-encoding in a string, panic if failure.
// This should only be used for hardcoded hash values.
func NewHashTFromStringAssert(data string) HashT {
	hash, err := NewHashTFromString(data)
	if err != nil {
		panic(err)
	}
	return hash
}

// Create a hash from a big.Int. Panics if data > 2^32-1
func NewHashTFromBigInt(data *big.Int) HashT {
	out := HashT{}
	data.FillBytes(out.data[:])
	return out
}

// Create a hash from a byte slice, panic if failure.
func NewHashTFromBytes(data []byte) HashT {
	if len(data) != 32 {
		panic(fmt.Sprintf("cannot create hash from %d bytes", len(data)))
	}
	return HashT{data: [32]byte(data)}
}

// Retrieve the underlying byte array from the HashT.
func (h HashT) Data() [32]byte {
	return h.data
}

// Convert to a hex-encoded string.
func (h HashT) String() string {
	return fmt.Sprintf("%x", h.data)
}

// Check whether this hash is equal in value to another.
func (h HashT) Eq(other HashT) bool {
	return h.data == other.data
}

// Check whether this hash is less than another (big-endian).
func (h HashT) Lt(other HashT) bool {
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

// Check whether this is the zero hash.
func (h HashT) EqZero() bool {
	return h.Eq(HashT{})
}

// Convert the Hash to a big.Int.
func (h HashT) BigInt() *big.Int {
	out := &big.Int{}
	out.SetBytes(h.data[:])
	return out
}

// Convert a hash target to a big.Int amount of work expected to beat it.
func (h HashT) TargetToWork() *big.Int {
	if h.data == [32]byte{} {
		panic("cannot compute work for zero target")
	}
	// return 2^256 / h.data
	out := h.BigInt()
	out.Div(bigInt2_256(), out)
	return out
}

// Increase the given total amount of work by the given target's work.
// Does not change hash value in-place.
func (h HashT) WorkAppendTarget(newTarget HashT) HashT {
	curInt := h.BigInt()
	nextInt := newTarget.TargetToWork()
	curInt.Add(curInt, nextInt)
	return NewHashTFromBigInt(curInt)
}

// Maximum (easiest) allowed next target after this one.
func (h HashT) MaxNextTarget(params Params) HashT {
	// Catch targets that are already out of bounds and bump
	if params.MaxTarget.Lt(h) {
		return params.MaxTarget
	}
	// Multiply h by 4
	hInt := h.BigInt()
	hInt.Mul(hInt, big.NewInt(4))
	newTarget := NewHashTFromBigInt(hInt)
	// If greater than max target, return max target
	if params.MaxTarget.Lt(newTarget) {
		return params.MaxTarget
	}
	return newTarget
}

// Minimum (hardest) allowed next target after this one.
func (h HashT) MinNextTarget() HashT {
	// Divide by 4
	hInt := h.BigInt()
	hInt.Div(hInt, big.NewInt(4))
	return NewHashTFromBigInt(hInt)
}

// Convert the given list of targets to an amount of total work to reach them all.
func TargetsToTotalWork(targets []HashT) *big.Int {
	total := &big.Int{}
	for _, target := range targets {
		total.Add(total, target.TargetToWork())
	}
	return total
}

// Compute 2^256-1 as a big.Int.
func bigInt2_256() *big.Int {
	out := &big.Int{}
	out.SetString(
		"ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
		16,
	)
	return out
}

func (h HashT) MarshalJSON() ([]byte, error) {
	return json.Marshal(h.String())
}

func (h *HashT) UnmarshalJSON(data []byte) error {
	var v string
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	parsed, err := NewHashTFromString(v)
	if err != nil {
		return err
	}
	h.data = parsed.data
	return nil
}

// Any object that defines how it is meant to be hashed.
type Hasher interface {
	Hash() HashT
}

// Generate a new double-sha256 hash from the given bytes.
func DHashBytes(content []byte) HashT {
	// Can't one-line bc [:] needs addressable memory
	first := sha256.Sum256(content)
	return HashT{
		data: sha256.Sum256(first[:]),
	}
}

// Generate a new double-sha256 hash from the given uint64.
func DHashUint64(content uint64) HashT {
	bs := make([]byte, 8)
	binary.BigEndian.PutUint64(bs, content)
	return DHashBytes(bs)
}

// Generate a new double-sha256 hash from the given bool.
func DHashBool(content bool) HashT {
	if content {
		return DHashBytes([]byte{0})
	} else {
		return DHashBytes([]byte{1})
	}
}

// Generate a new double-sha256 hash from whatever the given value is.
// If content is a hash, it's returned unchanged.
// If content is a Hasher, the output of its Hash() method is returned.
// If content is a uint64, the hash of its big-endian bytes is returned.
// If content is a byte slice, its hash is returned.
// If content is a bool, it's converted to a single 1 or 0 byte then hashed.
// If content is of unexpected type, this method panics.
func DHashAny(content any) HashT {
	switch typed := content.(type) {
	case Hasher:
		return typed.Hash()
	case HashT:
		return typed
	case uint64:
		return DHashUint64(typed)
	case []byte:
		return DHashBytes(typed)
	case bool:
		return DHashBool(typed)
	default:
		panic(fmt.Sprintf("unhashable type: %T", typed))
	}
}

// Generate a new double-sha256 hash of the given hashes, all concatenated.
func DHashHashes(items []HashT) HashT {
	concat := make([]byte, len(items)*32)
	for _, item := range items {
		concat = append(concat, item.data[:]...)
	}
	return DHashBytes(concat)
}

// Generate a new double-sha256 hash of the given various items concatenated.
// See DHashAny for documentation on how each type is handled
func DHashVarious(items ...any) HashT {
	hashes := make([]HashT, len(items))
	for i := range items {
		hashes[i] = DHashAny(items[i])
	}
	return DHashHashes(hashes)
}

func DHashList[T any](items []T) HashT {
	hashes := make([]HashT, len(items))
	for i := range items {
		hashes[i] = DHashAny(items[i])
	}
	return DHashHashes(hashes)
}
