package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/levilutz/basiccoin/src/db"
	"github.com/levilutz/basiccoin/src/util"
)

type Client struct {
	baseUrl string
}

// Create a new client from the given base url.
func NewClient(baseUrl string) (*Client, error) {
	if len(baseUrl) == 0 {
		return nil, fmt.Errorf("must provide client address")
	}
	if baseUrl[len(baseUrl)-1:] != "/" {
		baseUrl += "/"
	}
	c := &Client{
		baseUrl: baseUrl,
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

func (c *Client) SendTx(tx db.Tx) error {
	return nil
}

func (c *Client) GetHistory(publicKeyHashes ...db.HashT) []db.Tx {
	return []db.Tx{}
}
