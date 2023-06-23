package core

import "fmt"

// Various parameters that should be shared among all nodes in a network.
type Params struct {
	BlockReward      uint64 `json:"blockReward"`      // How much to reward the mining of a block.
	DifficultyPeriod uint64 `json:"difficultyPeriod"` // How many blocks between difficulty adjustments.
	BlockTargetTime  uint64 `json:"blockTargetTime"`  // Difficulty target for how long to mine a block.
	MaxBlockVSize    uint64 `json:"maxBlockVSize"`    // Maximum number of total hashed bytes in a block's txs.
	MaxTxVSize       uint64 `json:"maxTxVSize"`       // Maximum number of hashed bytes in a single tx.
	MaxTarget        HashT  `json:"maxTarget"`        // Maximum (easiest) allowed target value.
	OriginalTarget   HashT  `json:"originalTarget"`   // First block's required target difficulty
}

// Verify the parameters don't exceed limits.
func (p Params) verify() {
	// Verify MaxTarget below 3fff...
	// This ensures we can multiply by 4 safely
	maxAllowedMaxTarget := NewHashTFromStringAssert(
		"3fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
	)
	if maxAllowedMaxTarget.Lt(p.MaxTarget) {
		panic(fmt.Sprint("excessive max target:", p.MaxTarget))
	}
	// Verify DifficultyPeriod is at least 4
	// Lower values break computing difficulty adjustments
	if p.DifficultyPeriod < 4 {
		panic("difficulty period must be at least 4")
	}
}

// Generate params for the production network.
func ProdNetParams() Params {
	params := Params{
		BlockReward:      131072,  // 2^17 coin
		DifficultyPeriod: 128,     // 2^8 blocks
		BlockTargetTime:  720,     // 12 minutes
		MaxBlockVSize:    1048576, // 2^20 vBytes
		MaxTxVSize:       16384,   // 2^14 vBytes
		MaxTarget: NewHashTFromStringAssert(
			"0000000fffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
		), // 28 bits of 0s
		OriginalTarget: NewHashTFromStringAssert(
			"0000000fffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
		), // 28 bits of 0s
	}
	params.verify()
	return params
}

// Generate params for a local development network.
func DevNetParams() Params {
	params := Params{
		BlockReward:      1000,    // 1000 coin
		DifficultyPeriod: 8,       // 8 blocks
		BlockTargetTime:  10,      // 10 seconds
		MaxBlockVSize:    1048576, // 2^20 vBytes
		MaxTxVSize:       16384,   // 2^14 vBytes
		MaxTarget: NewHashTFromStringAssert(
			"000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
		), // 24 bits of 0s
		OriginalTarget: NewHashTFromStringAssert(
			"000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
		), // 24 bits of 0s
	}
	params.verify()
	return params
}
