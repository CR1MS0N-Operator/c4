// Package graphql provides a reusable GraphQL client for Mythic's Hasura-backed API.
package graphql

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client is a lightweight GraphQL client configured for Mythic's Hasura API.
type Client struct {
	endpoint   string
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a new GraphQL client.
//
//   - endpoint: full URL to the GraphQL endpoint, e.g. "https://127.0.0.1:7443/graphql/"
//   - hasuraSecret: the x-hasura-admin-secret value (empty string if not needed)
//   - skipVerify: set true to skip TLS certificate verification (self-signed certs)
//   - timeout: per-request timeout
func NewClient(endpoint, hasuraSecret string, skipVerify bool, timeout time.Duration) *Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: skipVerify, //nolint:gosec
		},
	}

	return &Client{
		endpoint: endpoint,
		apiKey:   hasuraSecret,
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   timeout,
		},
	}
}

// Do executes a GraphQL query and unmarshals the response into the provided value.
// The response value should be a pointer to a struct matching the GraphQL response shape.
func (c *Client) Do(ctx context.Context, query string, response any) error {
	body := map[string]string{"query": query}
	payload, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal query: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	if c.apiKey != "" {
		req.Header.Set("x-hasura-admin-secret", c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		errMsg := err.Error()
		switch {
		case strings.Contains(errMsg, "connection refused"):
			return fmt.Errorf("connection refused — is Mythic running at %s?", c.endpoint)
		case strings.Contains(errMsg, "no such host"):
			return fmt.Errorf("host unreachable — check config host %s", c.endpoint)
		case strings.Contains(errMsg, "timeout"):
			return fmt.Errorf("connection timed out — Mythic at %s is not responding", c.endpoint)
		default:
			return fmt.Errorf("http request: %w", err)
		}
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d from %s: %s", resp.StatusCode, c.endpoint, string(raw))
	}

	// Check for GraphQL-level errors
	var gqlResp struct {
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(raw, &gqlResp); err == nil && len(gqlResp.Errors) > 0 {
		msgs := make([]string, len(gqlResp.Errors))
		for i, e := range gqlResp.Errors {
			msgs[i] = e.Message
		}
		return fmt.Errorf("graphql error(s): %s", strings.Join(msgs, "; "))
	}

	if err := json.Unmarshal(raw, response); err != nil {
		return fmt.Errorf("unmarshal response: %w\nraw: %s", err, string(raw))
	}

	return nil
}

// Endpoint returns the configured GraphQL endpoint URL.
func (c *Client) Endpoint() string {
	return c.endpoint
}
