// Package loyverse provides a client for the Loyverse REST API v1.
// It covers items, inventory, receipts, shifts, and categories, with
// built-in cursor-based pagination and concurrent batch operations.
//
// Create a client with [New] and call endpoint methods directly:
//
//	client, err := loyverse.New(os.Getenv("LOYVERSE_TOKEN"))
//	if err != nil {
//	    log.Fatal(err)
//	}
//	items, err := client.GetItems(ctx)
package loyverse

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

const (
	_defaultBaseURL = "https://api.loyverse.com/v1.0"
	_defaultWorkers = 5
	_pageLimit      = "250"
)

// HTTPClient is the interface satisfied by the underlying HTTP transport.
// Override via [WithHTTPClient] to inject test doubles or custom transports.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client communicates with the Loyverse v1 API.
type Client struct {
	http    HTTPClient
	baseURL string
	token   string
	logger  *slog.Logger
	workers int
}

// Option configures a Client at construction time.
type Option func(*Client)

// WithHTTPClient replaces the default HTTP client.
// Use this in tests to inject an [httptest]-backed transport.
func WithHTTPClient(h HTTPClient) Option {
	return func(c *Client) { c.http = h }
}

// WithBaseURL overrides the API base URL.
// Primarily useful in tests with [httptest.NewServer].
func WithBaseURL(u string) Option {
	return func(c *Client) { c.baseURL = u }
}

// WithLogger sets the structured logger used for request tracing.
// Defaults to [slog.Default].
func WithLogger(l *slog.Logger) Option {
	return func(c *Client) { c.logger = l }
}

// WithBatchWorkers sets the goroutine concurrency for batch inventory operations.
// Must be a positive integer; values ≤ 0 are ignored. Defaults to 5.
func WithBatchWorkers(n int) Option {
	return func(c *Client) {
		if n > 0 {
			c.workers = n
		}
	}
}

// New creates a Loyverse API client authenticated with token.
// Returns an error if token is empty.
func New(token string, opts ...Option) (*Client, error) {
	if token == "" {
		return nil, fmt.Errorf("loyverse: token is required")
	}
	c := &Client{
		http:    newDefaultHTTPClient(),
		baseURL: _defaultBaseURL,
		logger:  slog.Default(),
		workers: _defaultWorkers,
		token:   token,
	}
	for _, o := range opts {
		o(c)
	}
	return c, nil
}

func newDefaultHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}
}

// get executes a GET request and decodes the JSON response into v.
func (c *Client) get(ctx context.Context, path string, params url.Values, v any) error {
	return c.do(ctx, http.MethodGet, path, params, nil, v)
}

// post marshals body as JSON, executes a POST, and decodes the response into v.
// v may be nil if the response body is not needed.
func (c *Client) post(ctx context.Context, path string, body, v any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("loyverse: marshal request body: %w", err)
	}
	return c.do(ctx, http.MethodPost, path, nil, bytes.NewReader(b), v)
}

// do is the core HTTP dispatch method used by all endpoint helpers.
func (c *Client) do(ctx context.Context, method, path string, params url.Values, body io.Reader, v any) error {
	fullURL := c.baseURL + path
	if len(params) > 0 {
		fullURL += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, body)
	if err != nil {
		return fmt.Errorf("loyverse: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	c.logger.DebugContext(ctx, "loyverse: request", "method", method, "url", fullURL)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("loyverse: execute request: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("loyverse: read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		c.logger.ErrorContext(ctx, "loyverse: API error",
			"status", resp.StatusCode,
			"url", fullURL,
		)
		return &APIError{StatusCode: resp.StatusCode, Body: string(respBytes)}
	}

	c.logger.DebugContext(ctx, "loyverse: response", "status", resp.StatusCode, "url", fullURL)

	if v == nil {
		return nil
	}
	return json.Unmarshal(respBytes, v)
}

// paginate executes cursor-based paginated fetches until no more pages remain.
// fetchPage is called with the current cursor and must return (items, nextCursor, error).
func paginate[T any](fetchPage func(cursor string) ([]T, string, error)) ([]T, error) {
	var all []T
	cursor := ""
	for {
		items, next, err := fetchPage(cursor)
		if err != nil {
			return nil, err
		}
		all = append(all, items...)
		if next == "" {
			break
		}
		cursor = next
	}
	return all, nil
}

// formatDate formats t as the UTC timestamp expected by Loyverse query parameters.
func formatDate(t time.Time) string {
	return t.UTC().Format("2006-01-02T15:04:05.000Z")
}
