package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/levilutz/basiccoin/internal/rest/models"
	"github.com/levilutz/basiccoin/pkg/core"
)

type WalletClient struct {
	rawUrl  string
	baseUrl string
	version string
}

func NewWalletClient(addr string, version string) (*WalletClient, error) {
	if len(addr) == 0 {
		return nil, fmt.Errorf("must provide node address")
	}
	rawUrl := addr
	if rawUrl[len(rawUrl)-1:] != "/" {
		rawUrl += "/"
	}
	baseUrl := rawUrl + "wallet/"
	c := &WalletClient{
		rawUrl:  rawUrl,
		baseUrl: baseUrl,
		version: version,
	}
	if err := c.Check(); err != nil {
		return nil, fmt.Errorf("failed to connect: %s", err.Error())
	}
	return c, nil
}

func (c *WalletClient) Check() error {
	resp, err := http.Get(c.rawUrl + "version")
	if err != nil {
		return err
	} else if resp.StatusCode != 200 {
		return fmt.Errorf("version non-2XX response: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	} else if string(body) != c.version {
		return fmt.Errorf("incompatible server version: '%s' != '%s'", string(body), c.version)
	}
	return nil
}

// Query the node for the current head height.
func (c *WalletClient) GetHeadHeight() (uint64, error) {
	resp, err := http.Get(c.baseUrl + "/head/height")
	if err != nil {
		return 0, err
	} else if resp.StatusCode != 200 {
		return 0, fmt.Errorf("/head/height received non-2XX response: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	return strconv.ParseUint(string(body), 10, 64)
}

// Query the node for the balance of a given pkh.
func (c *WalletClient) GetBalance(publicKeyHash core.HashT) (uint64, error) {
	queryStr := fmt.Sprintf("?publicKeyHash=%s", publicKeyHash)
	resp, err := GetParse[models.BalanceResp](c.baseUrl + "balance" + queryStr)
	if err != nil {
		return 0, err
	} else if _, ok := resp.Balances[publicKeyHash]; !ok {
		return 0, fmt.Errorf("did not receive correct pkhs in response")
	}
	return resp.Balances[publicKeyHash], nil
}

// Query the node for the balances of several pkhs.
func (c *WalletClient) GetManyBalances(publicKeyHashes []core.HashT) (map[core.HashT]uint64, error) {
	pkhStrs := core.MarshalHashTSlice(publicKeyHashes)
	queryStr := fmt.Sprintf("?publicKeyHash=%s", strings.Join(pkhStrs, "&publicKeyHash="))
	resp, err := GetParse[models.BalanceResp](c.baseUrl + "balance" + queryStr)
	if err != nil {
		return nil, err
	}
	return resp.Balances, nil
}

// Query the node for the utxos of multiple given pkhs.
func (c *WalletClient) GetManyUtxos(
	publicKeyHashes []core.HashT, excludeMempool bool,
) (map[core.Utxo]core.HashT, error) {
	pkhStrs := core.MarshalHashTSlice(publicKeyHashes)
	queryStr := fmt.Sprintf("?publicKeyHash=%s", strings.Join(pkhStrs, "&publicKeyHash="))
	if excludeMempool {
		queryStr += "&excludeMempool=true"
	}
	resp, err := GetParse[models.UtxosResp](c.baseUrl + "utxos" + queryStr)
	if err != nil {
		return nil, err
	}
	return resp.Utxos, nil
}

// Send a tx to the node.
func (c *WalletClient) PostTx(tx core.Tx) (core.HashT, error) {
	txJson, err := json.Marshal(tx)
	if err != nil {
		return core.HashT{}, err
	}
	resp, err := http.Post(c.baseUrl+"tx", "application/json", bytes.NewReader(txJson))
	if err != nil {
		return core.HashT{}, err
	} else if resp.StatusCode != 200 {
		return core.HashT{}, fmt.Errorf("tx non-2XX response: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return core.HashT{}, err
	}
	txId, err := core.NewHashTFromString(string(body))
	if err != nil {
		return core.HashT{}, fmt.Errorf("received txId failed to parse: %s - %s", txId, err.Error())
	}
	if txId != tx.Hash() {
		return core.HashT{}, fmt.Errorf("received incorrect txId: %s != %s", txId, tx.Hash())
	}
	return txId, nil
}

// Get tx confirmations.
func (c *WalletClient) GetTxConfirms(txIds []core.HashT) (map[core.HashT]uint64, error) {
	txIdStrs := core.MarshalHashTSlice(txIds)
	queryStr := fmt.Sprintf("?txId=%s", strings.Join(txIdStrs, "&txId="))
	resp, err := GetParse[models.TxConfirmsResp](c.baseUrl + "tx/confirms" + queryStr)
	if err != nil {
		return nil, err
	}
	return resp.Confirms, nil
}

// Get tx confirmations.
func (c *WalletClient) GetTxIncludedBlock(txIds []core.HashT) (map[core.HashT]core.HashT, error) {
	txIdStrs := core.MarshalHashTSlice(txIds)
	queryStr := fmt.Sprintf("?txId=%s", strings.Join(txIdStrs, "&txId="))
	resp, err := GetParse[models.TxIncludedBlockResp](c.baseUrl + "tx/block" + queryStr)
	if err != nil {
		return nil, err
	}
	return resp.IncludedBlocks, nil
}

// Get tx data.
func (c *WalletClient) GetTx(txIds []core.HashT) (map[core.HashT]core.Tx, error) {
	txIdStrs := core.MarshalHashTSlice(txIds)
	queryStr := fmt.Sprintf("?txId=%s", strings.Join(txIdStrs, "&txId="))
	resp, err := GetParse[models.GetTxResp](c.baseUrl + "tx" + queryStr)
	if err != nil {
		return nil, err
	}
	return resp.Txs, nil
}

// Get merkle data.
func (c *WalletClient) GetMerkle(merkleIds []core.HashT) (map[core.HashT]core.MerkleNode, error) {
	merkleIdStrs := core.MarshalHashTSlice(merkleIds)
	queryStr := fmt.Sprintf("?merkleId=%s", strings.Join(merkleIdStrs, "&merkleId="))
	resp, err := GetParse[models.GetMerkleResp](c.baseUrl + "merkle" + queryStr)
	if err != nil {
		return nil, err
	}
	return resp.Merkles, nil
}

// Get block data.
func (c *WalletClient) GetBlock(blockIds []core.HashT) (map[core.HashT]core.Block, error) {
	blockIdStrs := core.MarshalHashTSlice(blockIds)
	queryStr := fmt.Sprintf("?blockId=%s", strings.Join(blockIdStrs, "&blockId="))
	resp, err := GetParse[models.GetBlockResp](c.baseUrl + "block" + queryStr)
	if err != nil {
		return nil, err
	}
	return resp.Blocks, nil
}
