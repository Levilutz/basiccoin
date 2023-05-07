package utils

import (
	"encoding/json"
	"io"
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
