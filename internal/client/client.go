// Package client provides a GraphQL HTTP client.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Client is a GraphQL HTTP client.
type Client struct {
	endpoint   string
	httpClient *http.Client
	headers    map[string]string
}

// Request is a GraphQL request.
type Request struct {
	Query         string         `json:"query"`
	Variables     map[string]any `json:"variables,omitempty"`
	OperationName string         `json:"operationName,omitempty"`
}

// Response is a GraphQL response.
type Response struct {
	Data   json.RawMessage `json:"data,omitempty"`
	Errors []Error         `json:"errors,omitempty"`
}

// Error is a GraphQL error.
type Error struct {
	Message string `json:"message"`
}

// New creates a new client.
func New(endpoint string, opts ...Option) *Client {
	c := &Client{
		endpoint:   endpoint,
		httpClient: http.DefaultClient,
		headers:    make(map[string]string),
	}
	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Option configures the client.
type Option func(*Client)

// WithHeader adds a header.
func WithHeader(key, value string) Option {
	return func(c *Client) {
		c.headers[key] = value
	}
}

// Execute sends a request.
func (c *Client) Execute(ctx context.Context, req *Request) (*Response, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	for k, v := range c.headers {
		httpReq.Header.Set(k, v)
	}

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send: %w", err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("read: %w", err)
	}

	var resp Response
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	return &resp, nil
}
