package utils

import (
	"crypto/rand"
	"encoding/hex"
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
	AppVersion string
	RuntimeID  string
}

var Constants = ConstantsType{
	AppVersion: "0.1.0",
	RuntimeID:  AssertUUID(),
}
