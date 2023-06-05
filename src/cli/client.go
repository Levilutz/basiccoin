package main

import "github.com/levilutz/basiccoin/src/db"

type Client struct {
	Addr string
}

func (c *Client) CheckUrl(url string) bool {
	return false
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
