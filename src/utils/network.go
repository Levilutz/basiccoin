package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

func RetryGetBody[R any](url string, retries int) (body *R, midTimeMicro int64, err error) {
	var respBody *[]byte
	var last_err error
	for attempts := 0; attempts < retries; attempts++ {
		sentTime := time.Now().UnixMicro()
		resp, err := http.Get(url)
		respTime := time.Now().UnixMicro()
		midTimeMicro = (sentTime + respTime) / 2
		if err != nil {
			last_err = err
			continue
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			last_err = err
			continue
		}
		respBody = &body
		break
	}
	if respBody == nil {
		return nil, 0, last_err
	}
	var content R
	err = json.Unmarshal(*respBody, &content)
	if err != nil {
		return nil, 0, err
	}
	return &content, midTimeMicro, nil
}

func PostBody(url string, body any) error {
	reqBody, err := json.Marshal(body)
	if err != nil {
		return err
	}
	resp, err := http.Post(url, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return err
	}
	if resp.StatusCode >= 300 || resp.StatusCode < 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("status code %d != 2XX", resp.StatusCode)
		} else {
			return fmt.Errorf("status code %d != 2XX: %s", resp.StatusCode, body)
		}
	}
	return nil
}

func GetOutboundIP() (string, error) {
	cfAddr, err := net.ResolveTCPAddr("tcp", "1.1.1.1:80")
	if err != nil {
		return "", err
	}
	conn, err := net.DialTCP("tcp", nil, cfAddr)
	if err != nil {
		return "", err
	}
	defer conn.Close()
	return conn.LocalAddr().(*net.TCPAddr).IP.String(), nil
}
