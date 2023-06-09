package core

import (
	"crypto/ecdsa"
	"fmt"
	"sort"

	"github.com/levilutz/basiccoin/pkg/util"
)

// Manufacture a Tx using as few utxos as possible, and sending change back to wealthiest pkh.
// params is the core Params to use.
// privateKeys is a list of controlled private keys.
// utxoPkhs is a mapping of controlled utxos to their corresponding publicKeyHashes.
// dests is a mapping from destination pkhs to amounts to send to each.
// targetFeeRate is the goal fee rate in coin / vByte. (on average, this overshoots fees by 1 coin/output)
// minBlock is the minBlock to put on the tx.
func MakeOutboundTx(
	params Params,
	privateKeys []*ecdsa.PrivateKey,
	utxoPkhs map[Utxo]HashT,
	dests map[HashT]uint64,
	targetFeeRate float64,
	minBlock uint64,
) (*Tx, error) {
	// Make mapping from pkh to private keys
	pkhPrivs, err := getPkkPrivs(privateKeys)
	if err != nil {
		return nil, err
	}

	// Get total balance, pkh with the highest balance, and verify each pkh controlled
	balance, richestPkh, err := getBalanceAndRichestPkh(utxoPkhs, pkhPrivs)
	if err != nil {
		return nil, err
	}

	// Get total outputs and verify <= utxos
	totalOut := uint64(0)
	for _, val := range dests {
		totalOut += val
	}
	if totalOut >= balance {
		return nil, fmt.Errorf("insufficient balance: %d < %d", balance, totalOut)
	}

	// Get utxos sorted by value
	utxos := util.MapKeys(utxoPkhs)
	sort.Slice(utxos, func(i, j int) bool {
		// descending
		return utxos[i].Value > utxos[j].Value
	})

	// Build base tx
	tx := Tx{
		IsCoinbase: false,
		MinBlock:   minBlock,
		Inputs:     []TxIn{},
		Outputs:    make([]TxOut, len(dests)+1),
	}

	// Add placeholder change output, going to wealthiest controlled pkh
	tx.Outputs[0] = TxOut{
		Value:         0,
		PublicKeyHash: richestPkh,
	}

	// Add normal outputs
	i := 1
	for pkh, val := range dests {
		tx.Outputs[i] = TxOut{
			Value:         val,
			PublicKeyHash: pkh,
		}
		i++
	}

	// Add utxos, starting with the wealthiest, until we reach target input
	// Only using placeholder sigs, since preSigHash will change when we set change output value.
	totalIn := uint64(0)
	for i, utxo := range utxos {
		// Add the input
		totalIn += utxo.Value
		tx.Inputs = append(tx.Inputs, TxIn{
			Utxo:      utxo,
			PublicKey: ExamplePubDer(),
			Signature: ExampleMaxSigAsn(),
		})

		// Check if we just can't make a tx that fits within vSize limit
		if tx.VSize() > params.MaxTxVSize {
			return nil, fmt.Errorf("balance not consolidated enough - cannot create tx within vSize limits")
		}

		// Check if we now have enough inputs to cover outputs + fees
		if totalIn >= totalOut+tx.FeeFromRate(targetFeeRate) {
			break
		}

		// If we reached the end of the loop without breaking, we can't make it work
		if i == len(utxos)-1 {
			return nil, fmt.Errorf(
				"insufficient balance to pay outputs and target fee rate - %d < %d",
				balance,
				totalOut+tx.FeeFromRate(targetFeeRate),
			)
		}
	}

	// Set the change output
	// Ideally we would do this after replacing sigs bc vSize and thus fee would decrease
	// But unfortunately we need this output finalized so we can compute preSigHash
	// Thus we will on average overestimate vSize by ~1 vByte per output (<1% fee diff)
	tx.Outputs[0].Value = totalIn - totalOut - tx.FeeFromRate(targetFeeRate)

	// Sign the inputs, replacing placeholders
	preSigHash := TxHashPreSig(tx.MinBlock, tx.Outputs)
	for i := range tx.Inputs {
		utxo := utxos[i] // Don't range utxos as it's usually longer than tx.Inputs
		priv := pkhPrivs[utxoPkhs[utxo]]
		pub, err := MarshalEcdsaPublic(priv)
		if err != nil {
			return nil, err
		}
		sig, err := EcdsaSign(priv, preSigHash)
		if err != nil {
			return nil, err
		}
		tx.Inputs[i].PublicKey = pub
		tx.Inputs[i].Signature = sig
	}

	return &tx, nil
}

// Consolidate as much controlled balance as possible into a single utxo.
// params is the core Params to use.
// privateKeys is a list of controlled private keys.
// utxoPkhs is a mapping of controlled utxos to their corresponding publicKeyHashes.
// targetFeeRate is the goal fee rate in coin / vByte. (on average, this overshoots fee by 1 coin)
// minBlock is the minBlock to put on the tx.
func MakeConsolidateTx(
	params Params,
	privateKeys []*ecdsa.PrivateKey,
	utxoPkhs map[Utxo]HashT,
	targetFeeRate float64,
	minBlock uint64,
) (*Tx, error) {
	// Verify consolidation is possible
	if len(utxoPkhs) == 0 {
		return nil, fmt.Errorf("insufficient balance")
	} else if len(utxoPkhs) == 1 {
		return nil, fmt.Errorf("balance entirely consolidated")
	}

	// Make mapping from pkh to private keys
	pkhPrivs, err := getPkkPrivs(privateKeys)
	if err != nil {
		return nil, err
	}

	// Get pkh with the highest balance and verify each pkh controlled
	_, richestPkh, err := getBalanceAndRichestPkh(utxoPkhs, pkhPrivs)
	if err != nil {
		return nil, err
	}

	// Get utxos sorted by value
	utxos := util.MapKeys(utxoPkhs)
	sort.Slice(utxos, func(i, j int) bool {
		// descending
		return utxos[i].Value > utxos[j].Value
	})

	// Build base tx
	tx := Tx{
		IsCoinbase: false,
		MinBlock:   minBlock,
		Inputs:     []TxIn{},
		Outputs: []TxOut{
			{
				Value:         0,
				PublicKeyHash: richestPkh,
			},
		},
	}

	// Add utxos until we reach max size
	totalIn := uint64(0)
	for _, utxo := range utxos {
		txIn := TxIn{
			Utxo:      utxo,
			PublicKey: ExamplePubDer(),
			Signature: ExampleMaxSigAsn(),
		}
		if tx.VSize()+txIn.VSize() > params.MaxTxVSize {
			break
		}
		totalIn += utxo.Value
		tx.Inputs = append(tx.Inputs, txIn)
	}

	// Set the output amount
	tx.Outputs[0].Value = totalIn - tx.FeeFromRate(targetFeeRate)

	// Sign the inputs, replacing placeholders
	preSigHash := TxHashPreSig(tx.MinBlock, tx.Outputs)
	for i := range tx.Inputs {
		utxo := utxos[i] // Don't range utxos as it's usually longer than tx.Inputs
		priv := pkhPrivs[utxoPkhs[utxo]]
		pub, err := MarshalEcdsaPublic(priv)
		if err != nil {
			return nil, err
		}
		sig, err := EcdsaSign(priv, preSigHash)
		if err != nil {
			return nil, err
		}
		tx.Inputs[i].PublicKey = pub
		tx.Inputs[i].Signature = sig
	}

	return &tx, nil
}

// Given private keys, make a mapping from their public key hashes to each private key.
func getPkkPrivs(privateKeys []*ecdsa.PrivateKey) (map[HashT]*ecdsa.PrivateKey, error) {
	out := make(map[HashT]*ecdsa.PrivateKey, len(privateKeys))
	for _, priv := range privateKeys {
		pub, err := MarshalEcdsaPublic(priv)
		if err != nil {
			return nil, err
		}
		out[DHashBytes(pub)] = priv
	}
	return out, nil
}

// Get total balance, the pkh with the highest balance, and verify each pkh exists in given pkhPrivs.
func getBalanceAndRichestPkh(
	utxoPkhs map[Utxo]HashT, pkhPrivs map[HashT]*ecdsa.PrivateKey,
) (uint64, HashT, error) {
	balance := uint64(0)
	pkhBalances := make(map[HashT]uint64)
	for utxo, pkh := range utxoPkhs {
		if _, ok := pkhPrivs[pkh]; !ok {
			return 0, HashT{}, fmt.Errorf("pkh %s not controlled by private keys", pkh)
		}
		balance += utxo.Value
		pkhBalances[pkh] += utxo.Value
	}
	pkhs := util.MapKeys(pkhBalances)
	sort.Slice(pkhs, func(i, j int) bool {
		// descending
		return pkhBalances[pkhs[i]] > pkhBalances[pkhs[j]]
	})
	return balance, pkhs[0], nil
}
