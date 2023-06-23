package main

import (
	"fmt"
	"io"
	"net/http"
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
		baseUrl += "/"
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
	vers := getVersion(c.cfg.Dev)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	} else if string(body) != vers {
		return fmt.Errorf("incompatible server version: '%s' != '%s'", string(body), vers)
	}
	return nil
}

func getVersion(dev bool) string {
	if dev {
		return "v0.0.0-dev"
	} else {
		return "v0.0.0"
	}
}
