package main

import (
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"

	"github.com/levilutz/basiccoin/src/db"
	"github.com/levilutz/basiccoin/src/util"
)

type Client struct {
	baseUrl string
	config  *Config
}

// Data on the balances of several addressses.
type BalanceData struct {
	Balances    map[db.HashT]uint64
	Total       uint64
	SortedAddrs []db.HashT // Descending
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
func (c *Client) GetBalance(publicKeyHash db.HashT) (uint64, error) {
	queryStr := fmt.Sprintf("?publicKeyHash=%x", publicKeyHash)
	resp, err := http.Get(c.baseUrl + "balance" + queryStr)
	if err != nil {
		return 0, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	return strconv.ParseUint(string(body), 10, 64)
}

// Get balances per provided address and the total balance.
func (c *Client) GetBalances(pkhs []db.HashT) (*BalanceData, error) {
	out := make(map[db.HashT]uint64, len(pkhs))
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

// Send a tx to the node, return TxId.
func (c *Client) SendTx(tx db.Tx) (db.HashT, error) {
	return db.HashTZero, nil
}

// Create a tx given our current addresses.
func (c *Client) CreateOutboundTx(outputValues map[db.HashT]uint64) (db.Tx, error) {
	if len(outputValues) == 0 {
		return db.Tx{}, fmt.Errorf("no provided outputs")
	}
	// Get our current balances
	balanceData, err := c.GetBalances(c.config.GetPublicKeyHashes())
	if err != nil {
		return db.Tx{}, err
	}
	// Get total outputs
	totalOut := uint64(0)
	for _, val := range outputValues {
		totalOut += val
	}
	if totalOut > balanceData.Total {
		return db.Tx{}, fmt.Errorf("insufficient balance")
	}
	return db.Tx{}, nil
}

func (c *Client) GetHistory(publicKeyHashes ...db.HashT) []db.Tx {
	return []db.Tx{}
}
