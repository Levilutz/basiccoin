package main

import (
	"fmt"
	"net/http"

	"github.com/levilutz/basiccoin/src/db"
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
	_, err := http.Get(c.baseUrl + "version")
	if err != nil {
		return err
	}
	// fmt.Println(resp)
	return nil
}

func (c *Client) GetBalance(publicKeyHashes ...db.HashT) uint64 {
	return 0
}

func (c *Client) SendTx(tx db.Tx) error {
	return nil
}

func (c *Client) GetHistory(publicKeyHashes ...db.HashT) []db.Tx {
	return []db.Tx{}
}
