package utils

import (
	"encoding/json"
	"io"
	"net/http"
)

func RetryGetBody[R any](url string, retries int) (*R, error) {
	var resp_body *[]byte
	var last_err error = nil
	for attempts := 0; attempts < retries; attempts++ {
		resp, err := http.Get(url)
		if err != nil {
			last_err = err
			continue
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			last_err = err
			continue
		}
		resp_body = &body
		break
	}
	if resp_body == nil {
		return nil, last_err
	}
	var content R
	err := json.Unmarshal(*resp_body, &content)
	if err != nil {
		return nil, err
	}
	return &content, nil
}
