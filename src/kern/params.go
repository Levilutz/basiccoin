package kern

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

// Verify the MaxTarget isn't too high.
func (p Params) verifyMaxTarget() {
	// This ensures we can multiply by 4 safely
	maxAllowedMaxTarget, err := NewHashTFromString(
		"3fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
	)
	if err != nil {
		panic(err)
	}
	if maxAllowedMaxTarget.Lt(p.MaxTarget) {
		panic(fmt.Sprint("Excessive max target:", p.MaxTarget))
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
	params.verifyMaxTarget()
	return params
}

// Generate params for a local development network.
func DevNetParams() Params {
	params := Params{
		BlockReward:      1000,    // 1000 coin
		DifficultyPeriod: 32,      // 32 blocks
		BlockTargetTime:  15,      // 15 seconds
		MaxBlockVSize:    1048576, // 2^20 vBytes
		MaxTxVSize:       16384,   // 2^14 vBytes
		MaxTarget: NewHashTFromStringAssert(
			"000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
		), // 24 bits of 0s
		OriginalTarget: NewHashTFromStringAssert(
			"000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
		), // 24 bits of 0s
	}
	params.verifyMaxTarget()
	return params
}
