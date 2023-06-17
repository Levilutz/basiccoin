package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"

	"github.com/levilutz/basiccoin/src/kern"
	"github.com/levilutz/basiccoin/src/util"
)

type Client struct {
	baseUrl string
	config  *Config
}

// Data on the balances of several addressses.
type BalanceData struct {
	Balances    map[kern.HashT]uint64
	Total       uint64
	SortedAddrs []kern.HashT // Descending
}

// Create a new client from the given base url.
func NewClient(config *Config) (*Client, error) {
	if len(config.NodeAddr) == 0 {
		return nil, fmt.Errorf("must provide client address")
	}
	baseUrl := config.NodeAddr
	if config.NodeAddr[len(config.NodeAddr)-1:] != "/" {
		baseUrl += "/"
	}
	c := &Client{
		baseUrl: baseUrl,
		config:  config,
	}
	if err := c.Check(); err != nil {
		return nil, fmt.Errorf("client failed: %s", err.Error())
	}
	return c, nil
}

// Check that the server exists and is compatible with us.
func (c *Client) Check() error {
	resp, err := http.Get(c.baseUrl + "version")
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("version non-2XX response: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if string(body) != util.Constants.Version {
		return fmt.Errorf("incompatible server version '%s'", string(body))
	}
	return nil
}

// Query the node for the balance of the given address.
func (c *Client) GetBalance(publicKeyHash kern.HashT) (uint64, error) {
	queryStr := fmt.Sprintf("?publicKeyHash=%s", publicKeyHash)
	resp, err := http.Get(c.baseUrl + "balance" + queryStr)
	if err != nil {
		return 0, err
	}
	if resp.StatusCode != 200 {
		return 0, fmt.Errorf("balance non-2XX response: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	return strconv.ParseUint(string(body), 10, 64)
}

// Get balances per provided address and the total balance.
func (c *Client) GetAllBalances(pkhs []kern.HashT) (*BalanceData, error) {
	out := make(map[kern.HashT]uint64, len(pkhs))
	total := uint64(0)
	for _, pkh := range pkhs {
		bal, err := c.GetBalance(pkh)
		if err != nil {
			return nil, err
		}
		out[pkh] = bal
		total += bal
	}
	addrs := util.MapKeys(out)
	sort.Slice(addrs, func(i, j int) bool {
		// > instead of < because we want descending
		return out[addrs[i]] > out[addrs[j]]
	})
	return &BalanceData{
		Balances:    out,
		Total:       total,
		SortedAddrs: addrs,
	}, nil
}

// Query the node for the given address's utxos.
func (c *Client) GetUtxos(publicKeyHash kern.HashT) ([]kern.Utxo, error) {
	queryStr := fmt.Sprintf("?publicKeyHash=%s", publicKeyHash)
	resp, err := http.Get(c.baseUrl + "utxos" + queryStr)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("utxos non-2XX response: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	utxos := []kern.Utxo{}
	if err := json.Unmarshal(body, &utxos); err != nil {
		return nil, err
	}
	return utxos, nil
}

// Get utxos of all provided addresses. Return value maps utxos to their pkhs.
func (c *Client) GetAllUtxos(pkhs []kern.HashT) (map[kern.Utxo]kern.HashT, error) {
	out := make(map[kern.Utxo]kern.HashT)
	coveredPkhs := util.NewSet[kern.HashT]()
	for _, pkh := range pkhs {
		if coveredPkhs.Includes(pkh) {
			return nil, fmt.Errorf("duplicate pkh: %s", pkh)
		}
		coveredPkhs.Add(pkh)
		utxos, err := c.GetUtxos(pkh)
		if err != nil {
			return nil, err
		}
		for _, utxo := range utxos {
			if _, ok := out[utxo]; ok {
				return nil, fmt.Errorf("duplicate utxo: %s[%d]", utxo.TxId, utxo.Ind)
			}
			out[utxo] = pkh
		}
	}
	return out, nil
}

// Send a tx to the node, return TxId.
func (c *Client) SendTx(tx kern.Tx) (kern.HashT, error) {
	txJson, err := json.Marshal(tx)
	if err != nil {
		return kern.HashT{}, err
	}
	resp, err := http.Post(c.baseUrl+"tx", "application/json", bytes.NewReader(txJson))
	if err != nil {
		return kern.HashT{}, err
	}
	if resp.StatusCode != 200 {
		content, _ := io.ReadAll(resp.Body)
		fmt.Println(string(content))
		return kern.HashT{}, fmt.Errorf("tx non-2XX response: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return kern.HashT{}, err
	}
	txId, err := kern.NewHashTFromString(string(body))
	if err != nil {
		return kern.HashT{}, err
	}
	if txId != tx.Hash() {
		return kern.HashT{}, fmt.Errorf("wrong txId received: %s != %s", txId, tx.Hash())
	}
	return txId, nil
}

func (c *Client) GetHistory(publicKeyHashes ...kern.HashT) []kern.Tx {
	return []kern.Tx{}
}

// Manufacture an outbound tx that could be sent to the network.
// TODO: Allow this to be customized, don't use utxos that have unconfirmed spends.
func (c *Client) MakeOutboundTx(outputValues map[kern.HashT]uint64) (kern.Tx, error) {
	targetRate := float64(1.0) // Target fee rate in coin / vByte
	// Get available utxos
	utxos, err := c.GetAllUtxos(c.config.GetPublicKeyHashes())
	if err != nil {
		return kern.Tx{}, err
	} else if len(utxos) == 0 {
		return kern.Tx{}, fmt.Errorf("insufficient balance")
	}
	balance := uint64(0)
	pkhBalances := make(map[kern.HashT]uint64)
	for utxo, pkh := range utxos {
		balance += utxo.Value
		pkhBalances[pkh] += utxo.Value
	}
	sortedPkhs := util.MapKeys(pkhBalances)
	sort.Slice(sortedPkhs, func(i, j int) bool {
		// > instead of < because we want descending
		return pkhBalances[sortedPkhs[i]] > pkhBalances[sortedPkhs[j]]
	})

	// Get total outputs and verify <= utxos
	totalOut := uint64(0)
	for _, val := range outputValues {
		totalOut += val
	}
	if totalOut > balance {
		return kern.Tx{}, fmt.Errorf("insufficient balance")
	}

	// Get utxos sorted by value
	utxosSorted := util.MapKeys(utxos)
	sort.Slice(utxosSorted, func(i, j int) bool {
		// > instead of < because we want descending
		return utxosSorted[i].Value > utxosSorted[j].Value
	})

	// Build base tx with outputs and placeholder change output
	tx := kern.Tx{
		IsCoinbase: false,
		MinBlock:   0, // TODO: Query this from node
		Inputs:     []kern.TxIn{},
		Outputs: []kern.TxOut{
			{
				Value:         0,
				PublicKeyHash: sortedPkhs[0],
			},
		},
	}
	for pkh, val := range outputValues {
		tx.Outputs = append(tx.Outputs, kern.TxOut{
			Value:         val,
			PublicKeyHash: pkh,
		})
	}

	// Add utxos until we reach target input (with placeholder signatures)
	totalIn := uint64(0)
	for i, utxo := range utxosSorted {
		totalIn += utxo.Value

		tx.Inputs = append(tx.Inputs, kern.TxIn{
			OriginTxId:     utxo.TxId,
			OriginTxOutInd: utxo.Ind,
			PublicKey:      kern.ExamplePubDer(),
			Signature:      kern.ExampleMaxSigAsn(),
			Value:          utxo.Value,
		})
		if tx.VSize() > util.Constants.MaxTxVSize {
			return kern.Tx{}, fmt.Errorf("cannot create tx within vsize limits")
		}
		if totalIn >= totalOut+uint64(targetRate*float64(tx.VSize())) {
			break
		}
		if i == len(utxosSorted)-1 {
			return kern.Tx{}, fmt.Errorf("insufficient balance to pay target fee rate")
		}
	}

	// Set the change output
	// Ideally we would do this after replacing sigs bc vSize and thus fee would decrease
	// But unfortunately we need this output finalized so we can compute preSigHash
	// Thus we will on average overestimate vSize by ~1 vByte per output (<1% fee diff)
	tx.Outputs[0].Value = totalIn - totalOut - uint64(targetRate*float64(tx.VSize()))

	// Sign the inputs, replacing placeholders
	preSigHash := kern.TxHashPreSig(tx.MinBlock, tx.Outputs)
	for i := range tx.Inputs {
		utxo := utxosSorted[i]
		priv, err := c.config.GetPrivateKey(utxos[utxo])
		if err != nil {
			return kern.Tx{}, err
		}
		pub, err := kern.MarshalEcdsaPublic(priv)
		if err != nil {
			return kern.Tx{}, err
		}
		sig, err := kern.EcdsaSign(priv, preSigHash)
		if err != nil {
			return kern.Tx{}, err
		}
		tx.Inputs[i].PublicKey = pub
		tx.Inputs[i].Signature = sig
	}

	return tx, nil
}

// Manufacture a tx to consolidate as many of our utxos as possible.
func (c *Client) MakeConsolidateTx() (kern.Tx, error) {
	targetRate := float64(1.0) // Target fee rate coin / vByte
	// Get available utxos
	utxos, err := c.GetAllUtxos(c.config.GetPublicKeyHashes())
	if err != nil {
		return kern.Tx{}, err
	} else if len(utxos) == 0 {
		return kern.Tx{}, fmt.Errorf("insufficient balance")
	} else if len(utxos) == 1 {
		return kern.Tx{}, fmt.Errorf("no consolidation possible")
	}
	balance := uint64(0)
	pkhBalances := make(map[kern.HashT]uint64)
	for utxo, pkh := range utxos {
		balance += utxo.Value
		pkhBalances[pkh] += utxo.Value
	}
	sortedPkhs := util.MapKeys(pkhBalances)
	sort.Slice(sortedPkhs, func(i, j int) bool {
		// > instead of < because we want descending
		return pkhBalances[sortedPkhs[i]] > pkhBalances[sortedPkhs[j]]
	})

	// Get utxos sorted by value
	utxosSorted := util.MapKeys(utxos)
	sort.Slice(utxosSorted, func(i, j int) bool {
		// > instead of < because we want descending
		return utxosSorted[i].Value > utxosSorted[j].Value
	})

	// Build base tx with placeholder output
	tx := kern.Tx{
		IsCoinbase: false,
		MinBlock:   0, // TODO: Query this from the node
		Inputs:     []kern.TxIn{},
		Outputs: []kern.TxOut{
			{
				Value:         0,
				PublicKeyHash: sortedPkhs[0],
			},
		},
	}

	// Add utxos until we reach target (with placeholder sigs)
	totalIn := uint64(0)
	for _, utxo := range utxosSorted {
		newIn := kern.TxIn{
			OriginTxId:     utxo.TxId,
			OriginTxOutInd: utxo.Ind,
			PublicKey:      kern.ExamplePubDer(),
			Signature:      kern.ExampleMaxSigAsn(),
			Value:          utxo.Value,
		}
		if tx.VSize()+newIn.VSize() > util.Constants.MaxTxVSize {
			break
		}
		totalIn += utxo.Value
		tx.Inputs = append(tx.Inputs, newIn)
	}

	// Set the output amount
	tx.Outputs[0].Value = totalIn - uint64(targetRate*float64(tx.VSize()))

	// Sign the inputs, replacing placeholders
	preSigHash := kern.TxHashPreSig(tx.MinBlock, tx.Outputs)
	for i := range tx.Inputs {
		utxo := utxosSorted[i]
		priv, err := c.config.GetPrivateKey(utxos[utxo])
		if err != nil {
			return kern.Tx{}, err
		}
		pub, err := kern.MarshalEcdsaPublic(priv)
		if err != nil {
			return kern.Tx{}, err
		}
		sig, err := kern.EcdsaSign(priv, preSigHash)
		if err != nil {
			return kern.Tx{}, err
		}
		tx.Inputs[i].PublicKey = pub
		tx.Inputs[i].Signature = sig
	}

	return tx, nil
}
