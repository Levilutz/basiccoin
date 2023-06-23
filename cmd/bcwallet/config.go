package main

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path"

	"github.com/levilutz/basiccoin/pkg/core"
	"github.com/levilutz/basiccoin/src/kern"
)

type KeyConfigJSON struct {
	PublicKeyHash core.HashT `json:"publicKeyHash"`
	PrivateKey    []byte     `json:"privateKey"`
}

type KeyConfig struct {
	PublicKeyHash core.HashT
	PrivateKey    *ecdsa.PrivateKey
}

func NewKeyConfig(priv *ecdsa.PrivateKey) KeyConfig {
	pubBytes, err := kern.MarshalEcdsaPublic(priv)
	if err != nil {
		panic(err)
	}
	return KeyConfig{
		PublicKeyHash: core.DHashBytes(pubBytes),
		PrivateKey:    priv,
	}
}

func (kc KeyConfig) MarshalJSON() ([]byte, error) {
	privateBytes, err := core.MarshalEcdsaPrivate(kc.PrivateKey)
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(KeyConfigJSON{
		PublicKeyHash: kc.PublicKeyHash,
		PrivateKey:    privateBytes,
	}, "", "    ")
}

func (kc *KeyConfig) UnmarshalJSON(data []byte) error {
	v := KeyConfigJSON{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	priv, err := core.ParseECDSAPrivate(v.PrivateKey)
	if err != nil {
		return err
	}
	kc.PublicKeyHash = v.PublicKeyHash
	kc.PrivateKey = priv
	return nil
}

func (kc *KeyConfig) Verify() error {
	pub, err := core.MarshalEcdsaPublic(kc.PrivateKey)
	if err != nil {
		return err
	}
	if !core.DHashBytes(pub).Eq(kc.PublicKeyHash) {
		return fmt.Errorf("private key does not match public key hash")
	}
	return nil
}

type Config struct {
	Dev      bool        `json:"dev"`
	NodeAddr string      `json:"nodeAddr"`
	Keys     []KeyConfig `json:"keys"`
}

func NewConfig(nodeAddr string, dev bool) *Config {
	return &Config{
		Dev:      dev,
		NodeAddr: nodeAddr,
		Keys:     []KeyConfig{},
	}
}

func (c *Config) Version() string {
	if c.Dev {
		return "v0.0.0-dev"
	} else {
		return "v0.0.0"
	}
}

func (c *Config) CoreParams() core.Params {
	if c.Dev {
		return core.DevNetParams()
	} else {
		return core.ProdNetParams()
	}
}

func (c *Config) VerifyKeys() {
	for _, kc := range c.Keys {
		if err := kc.Verify(); err != nil {
			panic(err)
		}
	}
}

func (c *Config) HasPublicKeyHash(publicKeyHash core.HashT) bool {
	for _, kc := range c.Keys {
		if kc.PublicKeyHash == publicKeyHash {
			return true
		}
	}
	return false
}

func (c *Config) GetPublicKeyHashes() []core.HashT {
	out := make([]core.HashT, len(c.Keys))
	for i, kc := range c.Keys {
		out[i] = kc.PublicKeyHash
	}
	return out
}

func (c *Config) GetPrivateKeys() []*ecdsa.PrivateKey {
	out := make([]*ecdsa.PrivateKey, len(c.Keys))
	for i, kc := range c.Keys {
		out[i] = kc.PrivateKey
	}
	return out
}

func (c *Config) GetPrivateKey(publicKeyHash core.HashT) (*ecdsa.PrivateKey, error) {
	for _, kc := range c.Keys {
		if kc.PublicKeyHash == publicKeyHash {
			return kc.PrivateKey, nil
		}
	}
	return nil, fmt.Errorf("given public key hash not known by wallet")
}

func (c *Config) AddKeys(newKeys ...KeyConfig) {
	for _, kc := range newKeys {
		if !c.HasPublicKeyHash(kc.PublicKeyHash) {
			c.Keys = append(c.Keys, kc)
		}
	}
}

func getDefaultServer(dev bool) string {
	if dev {
		return "http://localhost:8080"
	} else {
		return "http://coin.levilutz.com:80"
	}
}

func getConfigDir() string {
	user, err := user.Current()
	if err != nil {
		panic("failed to get current user: " + err.Error())
	}
	return path.Join(user.HomeDir, ".config/basiccoin")
}

func getConfigPath(dev bool) string {
	if dev {
		return getConfigDir() + "/wallet-dev.json"
	} else {
		return getConfigDir() + "/wallet.json"
	}
}

func GetConfig(path string) *Config {
	rawConfig, err := os.ReadFile(path)
	if err != nil {
		panic("failed to find config: " + err.Error())
	}
	config := &Config{}
	if err := json.Unmarshal(rawConfig, config); err != nil {
		panic("failed to parse config: " + err.Error())
	}
	return config
}

func (c *Config) Save() error {
	rawConfig, err := json.MarshalIndent(c, "", "    ")
	if err != nil {
		return err
	}
	stat, err := os.Stat(getConfigDir())
	if os.IsNotExist(err) {
		os.MkdirAll(getConfigDir(), 0700)
	} else if err != nil {
		return err
	} else if !stat.IsDir() {
		return fmt.Errorf("config path is not a directory: %s", getConfigDir())
	}
	stat, err = os.Stat(getConfigPath(c.Dev))
	if err == nil && stat.IsDir() {
		return fmt.Errorf("config file is a directory: %s", getConfigPath(c.Dev))
	}

	return os.WriteFile(getConfigPath(c.Dev), rawConfig, 0600)
}

func EnsureConfig(dev bool) {
	_, err := os.ReadFile(getConfigPath(dev))
	if os.IsNotExist(err) {
		err = NewConfig(getDefaultServer(dev), dev).Save()
	}
	if err != nil {
		panic("failed to save config: " + err.Error())
	}
}
