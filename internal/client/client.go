// SPDX-License-Identifier: MPL-2.0

package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"path"
	"sync"

	"github.com/hashicorp/go-cleanhttp"
)

// Client is an Adaptive Logs API client.
type Client struct {
	Cfg     *Config
	BaseURL url.URL
	client  *http.Client

	// The Adaptive Logs API does not support concurrent writes to the
	// segments endpoint, so serialize access to it here.
	segmentMutex *sync.Mutex
}

// Config contains client configuration.
type Config struct {
	// APIKey is the "<tenant-id>:<token>" pair sent as a bearer token.
	APIKey      string
	HTTPHeaders map[string]string
	Debug       bool
	HttpClient  *http.Client

	UserAgent string
}

// New creates a new Adaptive Logs API client.
func New(baseURL string, cfg *Config) (*Client, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	if cfg.HttpClient == nil {
		cfg.HttpClient = cleanhttp.DefaultClient()
	}

	return &Client{
		Cfg:     cfg,
		BaseURL: *u,
		client:  cfg.HttpClient,

		segmentMutex: &sync.Mutex{},
	}, nil
}

func (c *Client) request(method, requestPath string, query url.Values, body []byte, responseStruct interface{}) error {
	log.Printf("request (%s) to %s with body data: %s", method, c.BaseURL.String(), string(body))
	req, err := c.newRequest(method, requestPath, query, bytes.NewReader(body))
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	bodyContents, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if c.Cfg.Debug {
		log.Printf("response status %d with body %v", resp.StatusCode, string(bodyContents))
	}

	switch {
	case resp.StatusCode == http.StatusNotFound:
		return ErrNotFound{BodyContents: bodyContents}
	case resp.StatusCode >= 400:
		return fmt.Errorf("%s", c.extractErrorMessage(bodyContents, resp.StatusCode))
	}

	if responseStruct == nil {
		return nil
	}

	return decodeResponse(bodyContents, responseStruct)
}

// decodeResponse unmarshals an API response into target, tolerating two
// envelope shapes the Adaptive Logs API mixes across endpoints: a bare
// value (e.g. segments, drop-rules) or a value wrapped as {"result": ...}
// (e.g. exemptions). It tries the bare shape first and only falls back to
// unwrapping "result" if that fails, so well-behaved bare responses never
// pay for the fallback.
func decodeResponse(bodyContents []byte, target interface{}) error {
	bareErr := json.Unmarshal(bodyContents, target)
	if bareErr == nil {
		return nil
	}

	var wrapper struct {
		Result json.RawMessage `json:"result"`
	}
	if err := json.Unmarshal(bodyContents, &wrapper); err != nil || wrapper.Result == nil {
		return bareErr
	}

	return json.Unmarshal(wrapper.Result, target)
}

func (c *Client) newRequest(method, requestPath string, query url.Values, body io.Reader) (*http.Request, error) {
	u := c.BaseURL
	u.Path = path.Join(u.Path, requestPath)
	u.RawQuery = query.Encode()
	req, err := http.NewRequest(method, u.String(), body)
	if err != nil {
		return req, err
	}

	if c.Cfg.APIKey != "" {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.Cfg.APIKey))
	}

	for k, v := range c.Cfg.HTTPHeaders {
		req.Header.Add(k, v)
	}

	req.Header.Add("User-Agent", c.Cfg.UserAgent)
	req.Header.Add("Content-Type", "application/json")
	return req, nil
}

func IsErrNotFound(err error) bool {
	var e ErrNotFound
	return errors.As(err, &e)
}

type ErrNotFound struct {
	BodyContents []byte
}

func (e ErrNotFound) Error() string {
	return fmt.Sprintf("status: 404, body: %s", e.BodyContents)
}

func (c *Client) extractErrorMessage(bodyContents []byte, statusCode int) string {
	var apiError struct {
		Error   string `json:"error"`
		Message string `json:"message"`
	}

	if err := json.Unmarshal(bodyContents, &apiError); err == nil {
		if msg := apiError.Error; msg != "" {
			return fmt.Sprintf("API error (status %d): %s", statusCode, msg)
		}
		if msg := apiError.Message; msg != "" {
			return fmt.Sprintf("API error (status %d): %s", statusCode, msg)
		}
	}

	return fmt.Sprintf("status: %d, body: %s", statusCode, string(bodyContents))
}
