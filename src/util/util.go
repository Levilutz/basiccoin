package util

import (
	"bufio"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"
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
	body := make([]byte, base64.StdEncoding.EncodedLen(len(bodyJson)))
	base64.StdEncoding.Encode(body, bodyJson)
	return body, nil
}

// Recovers unjson(unb64(body))
func UnJsonB64[R any](body []byte) (R, error) {
	var content R
	bodyJson := make([]byte, base64.StdEncoding.DecodedLen(len(body)))
	n, err := base64.StdEncoding.Decode(bodyJson, body)
	if err != nil {
		return content, err
	}
	err = json.Unmarshal(bodyJson[:n], &content)
	return content, err
}

// Pretty print json-able content
func PrettyPrint(content any) {
	b, err := json.MarshalIndent(content, "", "    ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}

// Retry reading a line from a bufio reader, exponential wait.
// Attempts begin at 1ms and multiply by 10.
// 5 Attempts is 11.111s total time, 6 attempts is 111.111s total time, etc.
func RetryReadLine(r *bufio.Reader, attempts int) ([]byte, error) {
	delay := time.Duration(1) * time.Millisecond
	for i := 0; i < attempts; i++ {
		data, err := r.ReadBytes(byte('\n'))
		if err == nil {
			if len(data) > 0 {
				return data[:len(data)-1], nil
			} else {
				return data, nil
			}
		} else if errors.Is(err, io.EOF) {
			time.Sleep(delay)
			delay *= time.Duration(10)
			continue
		} else {
			return nil, err
		}
	}
	return nil, io.EOF
}
