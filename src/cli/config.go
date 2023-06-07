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

// Verify that the public key hash matches the private key.
func (kc *KeyConfig) Verify() error {
	pub, err := db.MarshalEcdsaPublic(kc.PrivateKey)
	if err != nil {
		return err
	}
	if db.DHash(pub) != kc.PublicKeyHash {
		return fmt.Errorf("private key does not match public key hash")
	}
	return nil
}

type Config struct {
	NodeAddr string      `json:"nodeAddr"`
	Keys     []KeyConfig `json:"keys"`
}

func NewConfig(nodeAddr string) *Config {
	return &Config{
		NodeAddr: nodeAddr,
		Keys:     []KeyConfig{},
	}
}

func (cfg *Config) VerifyKeys() {
	for _, kc := range cfg.Keys {
		if err := kc.Verify(); err != nil {
			panic(err)
		}
	}
}

func getConfigDir() string {
	user, err := user.Current()
	if err != nil {
		panic("failed to get current user: " + err.Error())
	}
	return path.Join(user.HomeDir, ".config/basiccoin")
}

func getConfigPath() string {
	return getConfigDir() + "/cli.json"
}

// Get the current configuration, or nil if it doesn't exist.
func GetConfig(path string) *Config {
	rawConfig, err := os.ReadFile(getConfigPath())
	if err != nil {
		panic("failed to find config: " + err.Error())
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
	stat, err := os.Stat(getConfigDir())
	if os.IsNotExist(err) {
		os.MkdirAll(getConfigDir(), 0700)
	} else if err != nil {
		return err
	} else if !stat.IsDir() {
		return fmt.Errorf("config dir should be directory: %s", getConfigDir())
	}

	return os.WriteFile(getConfigPath(), rawConfig, 0600)
}

func EnsureConfig() {
	_, err := os.ReadFile(getConfigPath())
	if os.IsNotExist(err) {
		err = NewConfig("").Save()
	}
	if err != nil {
		panic("failed to save config: " + err.Error())
	}
}
