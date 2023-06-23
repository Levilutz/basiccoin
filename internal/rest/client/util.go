package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func GetParse[K json.Marshaler](url string) (out K, err error) {
	resp, err := http.Get(url)
	if err != nil {
		return out, err
	} else if resp.StatusCode != 200 {
		return out, fmt.Errorf("%s non-2XX response: %d", url, resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return out, err
	}
	if err = json.Unmarshal(body, &out); err != nil {
		return out, err
	}
	return out, nil
}
