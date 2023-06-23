package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/levilutz/basiccoin/pkg/core"
)

type Client struct {
	baseUrl string
	cfg     *Config
}

func NewClient(cfg *Config) (*Client, error) {
	if len(cfg.NodeAddr) == 0 {
		return nil, fmt.Errorf("must provide client address")
	}
	baseUrl := cfg.NodeAddr
	if baseUrl[len(baseUrl)-1:] != "/" {
		baseUrl += "/wallet/"
	} else {
		baseUrl += "wallet/"
	}
	c := &Client{
		baseUrl: baseUrl,
		cfg:     cfg,
	}
	if err := c.Check(); err != nil {
		return nil, fmt.Errorf("client failed: %s", err.Error())
	}
	return c, nil
}

func (c *Client) Check() error {
	resp, err := http.Get(c.baseUrl + "version")
	if err != nil {
		return err
	} else if resp.StatusCode != 200 {
		return fmt.Errorf("version non-2XX response: %d", resp.StatusCode)
	}
	vers := c.cfg.Version()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	} else if string(body) != vers {
		return fmt.Errorf("incompatible server version: '%s' != '%s'", string(body), vers)
	}
	return nil
}

// Query the node for the balance of a given pkh.
func (c *Client) GetBalance(publicKeyHash core.HashT) (uint64, error) {
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
