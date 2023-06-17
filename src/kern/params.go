package kern

// Various parameters that should be shared among all nodes in a network.
type Params struct {
	BlockReward      uint64 // How much to reward the mining of a block.
	DifficultyPeriod uint64 // How many blocks between difficulty adjustments.
	BlockTargetTime  uint64 // Difficulty target for how long to mine a block.
	MaxBlockVSize    uint64 // Maximum number of total hashed bytes in a block's txs.
	MaxTxVSize       uint64 // Maximum number of hashed bytes in a single tx.
}

// Generate params for the production network.
func ProdNetParams() Params {
	return Params{
		BlockReward:      131072,  // 2^17 coin
		DifficultyPeriod: 128,     // 2^8 blocks
		BlockTargetTime:  720,     // 12 minutes
		MaxBlockVSize:    1048576, // 2^20 vBytes
		MaxTxVSize:       16384,   // 2^14 vBytes
	}
}

// Generate params for a local development network.
func DevNetParams() Params {
	return Params{
		BlockReward:      1000,    // 1000 coin
		DifficultyPeriod: 32,      // 32 blocks
		BlockTargetTime:  15,      // 15 seconds
		MaxBlockVSize:    1048576, // 2^20 vBytes
		MaxTxVSize:       16384,   // 2^14 vBytes
	}
}
