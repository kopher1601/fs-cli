package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	defaultBaseURL = "https://generativelanguage.googleapis.com/v1beta"
	uploadBaseURL  = "https://generativelanguage.googleapis.com/upload/v1beta"
)

// Client is the HTTP client for the Gemini File Search API.
type Client struct {
	baseURL    string
	uploadURL  string
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a new API client with the given API key.
func NewClient(apiKey string) *Client {
	return &Client{
		baseURL:   defaultBaseURL,
		uploadURL: uploadBaseURL,
		apiKey:    apiKey,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

func (c *Client) buildURL(base, path string, params url.Values) string {
	if params == nil {
		params = url.Values{}
	}
	params.Set("key", c.apiKey)
	return fmt.Sprintf("%s/%s?%s", base, path, params.Encode())
}

func (c *Client) doRequest(ctx context.Context, method, reqURL string, body interface{}, out interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL, bodyReader)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var apiErr APIError
		if json.Unmarshal(respBody, &apiErr) == nil && apiErr.Error.Message != "" {
			return fmt.Errorf("API error %d (%s): %s", apiErr.Error.Code, apiErr.Error.Status, apiErr.Error.Message)
		}
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
	}

	if out != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, out); err != nil {
			return fmt.Errorf("unmarshal response: %w", err)
		}
	}

	return nil
}

func (c *Client) get(ctx context.Context, path string, params url.Values, out interface{}) error {
	return c.doRequest(ctx, http.MethodGet, c.buildURL(c.baseURL, path, params), nil, out)
}

func (c *Client) post(ctx context.Context, path string, body interface{}, out interface{}) error {
	return c.doRequest(ctx, http.MethodPost, c.buildURL(c.baseURL, path, nil), body, out)
}

func (c *Client) delete(ctx context.Context, path string, params url.Values) error {
	return c.doRequest(ctx, http.MethodDelete, c.buildURL(c.baseURL, path, params), nil, nil)
}
