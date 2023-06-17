package db

import (
	"crypto/rand"
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

func (h HashT2) Data() [32]byte {
	return h.data
}

func (h HashT2) String() string {
	return fmt.Sprintf("%x", h.data)
}

func (h HashT2) Eq(other HashT2) bool {
	return h.data == other.data
}

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

func (h HashT2) BigInt() *big.Int {
	out := &big.Int{}
	out.SetBytes(h.data[:])
	return out
}

func (h HashT2) TargetToWork() *big.Int {
	if h.data == [32]byte{} {
		panic("cannot compute work for zero target")
	}
	// return 2^256 / h.data
	out := h.BigInt()
	out.Div(bigInt2_256(), out)
	return out
}

func (h HashT2) WorkAppendTarget(newTarget HashT2) HashT2 {
	curInt := h.BigInt()
	nextInt := newTarget.TargetToWork()
	curInt.Add(curInt, nextInt)
	return NewHashT2FromBigInt(curInt)
}

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
