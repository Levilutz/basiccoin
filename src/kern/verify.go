package kern

import (
	"fmt"

	"github.com/levilutz/basiccoin/src/util"
)

// Verify what we can about this transaction in isolation.
func (tx Tx) VerifyIsolated() error {
	preSigHash := TxHashPreSig(tx.MinBlock, tx.Outputs)
	for _, txi := range tx.Inputs {
		valid, err := EcdsaVerify(txi.PublicKey, preSigHash, txi.Signature)
		if err != nil || !valid {
			return fmt.Errorf("tx signature invalid")
		}
	}
	if tx.VSize() > util.Constants.MaxTxVSize {
		return fmt.Errorf("tx vSize exceeds limit")
	}
	if tx.IsCoinbase {
		if len(tx.Inputs) > 0 {
			return fmt.Errorf("coinbase cannot have inputs")
		} else if len(tx.Outputs) != 1 {
			return fmt.Errorf("coinbase must have 1 output")
		} else if tx.OutputsValue() < util.Constants.BlockReward {
			return fmt.Errorf("coinbase has insufficient block reward")
		}
	} else {
		if tx.OutputsValue() >= tx.InputsValue() {
			return fmt.Errorf("tx outputs exceed or match inputs")
		}
	}
	return nil
}

// Verify what we can about this block in isolation.
func (b Block) VerifyIsolated() error {
	if !b.Hash().Lt(b.Difficulty) {
		return fmt.Errorf("block fails to beat claimed target difficulty")
	}
	return nil
}
