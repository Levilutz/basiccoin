package db

import (
	"fmt"

	"github.com/levilutz/basiccoin/src/util"
)

// Verify a block.
func (s *State) VerifyExistingBlock(blockId HashT, txIds []HashT) error {
	// Verify block exists
	bh, exists := s.ledger[blockId]
	if !exists {
		return fmt.Errorf("unknown blockId: %s", blockId)
	}

	if err := bh.Verify(); err != nil {
		return err
	}

	// Verify each claimed txId is known, retrieve txs
	txs := make([]Tx, len(txIds))
	for i, txId := range txIds {
		tx, ok := s.txs[txId]
		if !ok {
			return fmt.Errorf("unknown txId: %s", txId)
		}
		txs[i] = tx
	}

	// TODO: Verify no duplicate UTxOs (helper?)

	// TODO: Verify total vSize within bounds

	return nil
}

func (s *State) VerifyNewBlock(bh BlockHeader, txIds []HashT) error {
	if err := bh.Verify(); err != nil {
		return err
	}

	// Verify each claimed txId is known, retrieve txs
	txs := make([]Tx, len(txIds))
	for i, txId := range txIds {
		tx, ok := s.txs[txId]
		if !ok {
			return fmt.Errorf("unknown txId: %s", txId)
		}
		if _, existsMempool := s.mempool[txId]; !existsMempool {
			return fmt.Errorf("txId not in mempool: %s", txId)
		}
		txs[i] = tx
	}

	consumedUTxOs := make(map[UTxO]struct{}, 0)
	for i, tx := range txs {
		// Verify the Tx
		err := s.VerifyTx(tx, i == 0)
		if err != nil {
			return fmt.Errorf("tx failed verification: %s", err.Error())
		}

		// Verify no double-spent UTxOs in transaction set
		for _, txi := range tx.Inputs {
			ClaimedUTxO := UTxO{
				TxId: txi.OriginTxId,
				Ind:  txi.OriginTxOutInd,
			}
			if _, exists := consumedUTxOs[ClaimedUTxO]; exists {
				return fmt.Errorf(
					"double-spend on UTxO %s[%d]", txi.OriginTxId, txi.OriginTxOutInd,
				)
			}
			consumedUTxOs[ClaimedUTxO] = struct{}{}
		}
	}

	// TODO: Verify total vSize within bounds

	// TODO: Verify claimed target difficulty matches ours
	return nil
}

// Verify a transaction and compute auxiliary information.
func (s *State) VerifyTx(tx Tx, coinbaseExpected bool) error {
	preSigHash := TxHashPreSig(tx.MinBlock, tx.Outputs)

	var totalInput uint64 = 0
	var vSize uint64 = 0
	refOutputs := make([]TxOut, len(tx.Inputs))

	if coinbaseExpected && len(tx.Inputs) != 0 {
		return fmt.Errorf("coinbase tx cannot have inputs")
	} else if !coinbaseExpected && len(tx.Inputs) == 0 {
		return fmt.Errorf("non-coinbase tx must have inputs")
	}

	for i, txi := range tx.Inputs {
		// Verify the claimed UTxO exists
		ClaimedUTxO := UTxO{
			TxId: txi.OriginTxId,
			Ind:  txi.OriginTxOutInd,
		}
		_, exists := s.uTxOs[ClaimedUTxO]
		if !exists {
			return fmt.Errorf(
				"invalid claimed UTxO %s[%d]",
				txi.OriginTxId,
				txi.OriginTxOutInd,
			)
		}

		// Retrieve referenced output from ledger
		refOutputs[i] = s.txs[txi.OriginTxId].Outputs[txi.OriginTxOutInd]

		// Verify claimed public key matches
		claimedPubKeyHash := DHash(txi.PublicKey)
		if claimedPubKeyHash != refOutputs[i].PublicKeyHash {
			return fmt.Errorf(
				"invalid claimed public key hash %s", claimedPubKeyHash,
			)
		}

		// Verify signature matches
		valid, err := EcdsaVerify(txi.PublicKey, preSigHash, txi.Signature)
		if err != nil {
			return err
		} else if !valid {
			return fmt.Errorf("invalid signature")
		}

		// Tally input value and vSize
		totalInput += uint64(refOutputs[i].Value)
		vSize += uint64(len(txi.PublicKey))
		vSize += uint64(len(txi.Signature))
	}

	// Tally output value and more vSize
	var totalOutput uint64 = 0
	for _, output := range tx.Outputs {
		totalOutput += uint64(output.Value)
		vSize += uint64(len(output.PublicKeyHash))
	}

	// Verify no more outputs than inputs
	if totalOutput > totalInput {
		return fmt.Errorf(
			"outputs exceed inputs: %d > %d", totalOutput, totalInput,
		)
	}

	// Verify vSize is within limits
	if vSize > util.Constants.MaxVSize {
		return fmt.Errorf("transaction too large: %d vB", vSize)
	}

	return nil
}
