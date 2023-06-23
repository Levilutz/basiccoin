package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/levilutz/basiccoin/pkg/core"
	"github.com/levilutz/basiccoin/pkg/set"
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
	resp, err := http.Get(c.baseUrl + "balance" + queryStr)
	if err != nil {
		return 0, err
	} else if resp.StatusCode != 200 {
		return 0, fmt.Errorf("balance non-2XX response: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	return strconv.ParseUint(string(body), 10, 64)
}

// Query the node for the balances of several pkhs.
func (c *WalletClient) GetManyBalances(publicKeyHashes []core.HashT) (map[core.HashT]uint64, error) {
	balances := make(map[core.HashT]uint64, len(publicKeyHashes))
	for _, pkh := range publicKeyHashes {
		bal, err := c.GetBalance(pkh)
		if err != nil {
			return nil, err
		}
		balances[pkh] = bal
	}
	return balances, nil
}

// Query the node for the utxos of a given pkh.
func (c *WalletClient) GetUtxos(publicKeyHash core.HashT) ([]core.Utxo, error) {
	queryStr := fmt.Sprintf("?publicKeyHash=%s", publicKeyHash)
	resp, err := http.Get(c.baseUrl + "utxos" + queryStr)
	if err != nil {
		return nil, err
	} else if resp.StatusCode != 200 {
		return nil, fmt.Errorf("utxos non-2XX response: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	utxos := []core.Utxo{}
	if err := json.Unmarshal(body, &utxos); err != nil {
		return nil, err
	}
	return utxos, nil
}

// Query the node for the utxos of multiple given pkhs.
func (c *WalletClient) GetManyUtxos(publicKeyHashes []core.HashT) (map[core.Utxo]core.HashT, error) {
	out := make(map[core.Utxo]core.HashT)
	covered := set.NewSet[core.HashT]()
	for _, pkh := range publicKeyHashes {
		if covered.Includes(pkh) {
			continue
		}
		covered.Add(pkh)
		utxos, err := c.GetUtxos(pkh)
		if err != nil {
			return nil, err
		}
		for _, utxo := range utxos {
			if _, ok := out[utxo]; ok {
				continue
			}
			out[utxo] = pkh
		}
	}
	return out, nil
}
