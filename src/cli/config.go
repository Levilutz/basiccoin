package main

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path"

	"github.com/levilutz/basiccoin/src/db"
)

type KeyConfigJSON struct {
	PublicKeyHash string `json:"publicKeyHash"`
	PrivateKey    []byte `json:"privateKey"`
}

type KeyConfig struct {
	PublicKeyHash db.HashT
	PrivateKey    *ecdsa.PrivateKey
}

func (kc KeyConfig) MarshalJSON() ([]byte, error) {
	privateBytes, err := db.MarshalEcdsaPrivate(kc.PrivateKey)
	if err != nil {
		return nil, err
	}
	return json.Marshal(KeyConfigJSON{
		PublicKeyHash: fmt.Sprintf("%x", kc.PublicKeyHash),
		PrivateKey:    privateBytes,
	})
}

func (kc *KeyConfig) UnmarshalJSON(data []byte) error {
	v := KeyConfigJSON{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	pkh, err := db.StringToHash(v.PublicKeyHash)
	if err != nil {
		return err
	}
	priv, err := db.ParseECDSAPrivate(v.PrivateKey)
	if err != nil {
		return err
	}
	kc.PublicKeyHash = pkh
	kc.PrivateKey = priv
	return nil
}

type Config struct {
	NodeAddr string      `json:"nodeAddr"`
	Keys     []KeyConfig `json:"keys"`
}

func getConfigPath() string {
	user, err := user.Current()
	if err != nil {
		panic("failed to get current user: " + err.Error())
	}
	return path.Join(user.HomeDir, ".config/basiccoin/cli.json")
}

// Get the current configuration, or nil if it doesn't exist.
func GetConfig() *Config {
	rawConfig, err := os.ReadFile(getConfigPath())
	if err != nil {
		return nil
	}
	config := Config{}
	if err := json.Unmarshal(rawConfig, &config); err != nil {
		panic("failed to parse config: " + err.Error())
	}
	return &config
}

// Save the configuration.
func (cfg *Config) Save() error {
	rawConfig, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(getConfigPath(), rawConfig, 0600)
}