// Package httpx extends the stdlib http client with some opinionated defaults
// and utility methods.
package httpx

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	*http.Client
}

func NewClient() *Client {
	return &Client{
		Client: &http.Client{
			Timeout: time.Second * 30,
		},
	}
}

func (c *Client) GetJson(ctx context.Context, url string, v interface{}) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("preparing http request: %w", err)
	}

	res, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("during http request: %w", err)
	}
	defer res.Body.Close()

	return res, json.NewDecoder(res.Body).Decode(v)
}
