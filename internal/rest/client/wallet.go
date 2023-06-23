package client

import (
	"fmt"
	"io"
	"net/http"
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
	pkhStrs := make([]string, len(publicKeyHashes))
	for i, pkh := range publicKeyHashes {
		pkhStrs[i] = pkh.String()
	}
	queryStr := fmt.Sprintf("?publicKeyHash=%s", strings.Join(pkhStrs, "&publicKeyHash="))
	resp, err := GetParse[models.BalanceResp](c.baseUrl + "balance" + queryStr)
	if err != nil {
		return nil, err
	}
	return resp.Balances, nil
}

// Query the node for the utxos of a given pkh.
func (c *WalletClient) GetUtxos(publicKeyHash core.HashT) ([]core.Utxo, error) {
	queryStr := fmt.Sprintf("?publicKeyHash=%s", publicKeyHash)
	resp, err := GetParse[models.UtxosResp](c.baseUrl + "utxos" + queryStr)
	if err != nil {
		return nil, err
	}
	out := make([]core.Utxo, len(resp.Utxos))
	i := 0
	for utxo, pkh := range resp.Utxos {
		if pkh != publicKeyHash {
			return nil, fmt.Errorf("did not receive correct pkh in response")
		}
		out[i] = utxo
		i++
	}
	return out, nil
}

// Query the node for the utxos of multiple given pkhs.
func (c *WalletClient) GetManyUtxos(publicKeyHashes []core.HashT) (map[core.Utxo]core.HashT, error) {
	pkhStrs := make([]string, len(publicKeyHashes))
	for i, pkh := range publicKeyHashes {
		pkhStrs[i] = pkh.String()
	}
	queryStr := fmt.Sprintf("?publicKeyHash=%s", strings.Join(pkhStrs, "&publicKeyHash="))
	resp, err := GetParse[models.UtxosResp](c.baseUrl + "utxos" + queryStr)
	if err != nil {
		return nil, err
	}
	return resp.Utxos, nil
}
