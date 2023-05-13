package util

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// Generate UUID
func UUID() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// Generate UUID, panic on failure
func AssertUUID() string {
	uuid, err := UUID()
	if err != nil {
		panic(err)
	}
	return uuid
}

// Generates base64(json(content))
func JsonB64(content any) ([]byte, error) {
	bodyJson, err := json.Marshal(content)
	if err != nil {
		return []byte{}, err
	}
	body := make([]byte, 0)
	base64.StdEncoding.Encode(body, bodyJson)
	return body, nil
}

// Pretty print json-able content
func PrettyPrint(content any) {
	b, err := json.MarshalIndent(content, "", "    ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}
