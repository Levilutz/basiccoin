package utils

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
)

func UUID() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func AssertUUID() string {
	uuid, err := UUID()
	if err != nil {
		panic(err)
	}
	return uuid
}

type ConstantsType struct {
	RuntimeID string `json:"runtimeID"`
	Version   string `json:"version"`
}

func (ct ConstantsType) AsJSON() string {
	b, err := json.MarshalIndent(ct, "", "    ")
	if err != nil {
		panic(err)
	}
	return string(b)
}

var Constants = ConstantsType{
	RuntimeID: AssertUUID(),
	Version:   "0.1.0",
}
