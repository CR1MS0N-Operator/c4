// Package graphql provides a reusable GraphQL client for Mythic's Hasura-backed API.
package graphql

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// Client is a lightweight GraphQL client configured for Mythic's Hasura API.
type Client struct {
	endpoint   string
	apiKey     string
	httpClient *http.Client
}

// TLSOptions controls how the server certificate of an https endpoint is
// verified. The zero value verifies against the system trust store. At most
// one field may be set.
type TLSOptions struct {
	// CAFile is a PEM bundle that replaces the system trust store.
	CAFile string
	// CertFingerprint pins the server's leaf certificate by its SHA-256
	// fingerprint (hex, optionally colon-separated or "sha256:"-prefixed).
	CertFingerprint string
	// InsecureSkipVerify disables certificate verification entirely.
	InsecureSkipVerify bool
}

// NewClient creates a new GraphQL client.
//
//   - endpoint: full URL to the GraphQL endpoint, e.g. "https://127.0.0.1:7443/graphql/"
//   - hasuraSecret: the x-hasura-admin-secret value (empty string if not needed)
//   - tlsOpts: how to verify the server certificate
//   - timeout: per-request timeout
func NewClient(endpoint, hasuraSecret string, tlsOpts TLSOptions, timeout time.Duration) (*Client, error) {
	tlsConfig, err := tlsOpts.config()
	if err != nil {
		return nil, err
	}

	return &Client{
		endpoint: endpoint,
		apiKey:   hasuraSecret,
		httpClient: &http.Client{
			Transport:     &http.Transport{TLSClientConfig: tlsConfig},
			Timeout:       timeout,
			CheckRedirect: sameOriginRedirect,
		},
	}, nil
}

// sameOriginRedirect refuses redirects that leave the original scheme and
// host, so the admin secret header is never forwarded to another origin.
func sameOriginRedirect(req *http.Request, via []*http.Request) error {
	if len(via) >= 10 {
		return errors.New("stopped after 10 redirects")
	}
	orig := via[0].URL
	if req.URL.Scheme != orig.Scheme || req.URL.Host != orig.Host {
		return fmt.Errorf("refusing redirect from %s://%s to %s://%s", orig.Scheme, orig.Host, req.URL.Scheme, req.URL.Host)
	}
	return nil
}

func (o TLSOptions) config() (*tls.Config, error) {
	set := 0
	for _, on := range []bool{o.CAFile != "", o.CertFingerprint != "", o.InsecureSkipVerify} {
		if on {
			set++
		}
	}
	if set > 1 {
		return nil, errors.New("tls: set only one of ca_file, cert_fingerprint, insecure_skip_verify")
	}

	cfg := &tls.Config{MinVersion: tls.VersionTLS12}
	switch {
	case o.CAFile != "":
		pem, err := os.ReadFile(o.CAFile)
		if err != nil {
			return nil, fmt.Errorf("tls: read ca_file: %w", err)
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(pem) {
			return nil, fmt.Errorf("tls: ca_file %q contains no PEM certificates", o.CAFile)
		}
		cfg.RootCAs = pool
	case o.CertFingerprint != "":
		want, err := parseFingerprint(o.CertFingerprint)
		if err != nil {
			return nil, err
		}
		// Chain and hostname checks are replaced by an exact match on the leaf.
		cfg.InsecureSkipVerify = true //nolint:gosec
		cfg.VerifyConnection = func(cs tls.ConnectionState) error {
			if len(cs.PeerCertificates) == 0 {
				return errPinMismatch
			}
			got := sha256.Sum256(cs.PeerCertificates[0].Raw)
			if subtle.ConstantTimeCompare(got[:], want) != 1 {
				return errPinMismatch
			}
			return nil
		}
	case o.InsecureSkipVerify:
		cfg.InsecureSkipVerify = true //nolint:gosec
	}
	return cfg, nil
}

var errPinMismatch = errors.New("server certificate does not match cert_fingerprint")

func parseFingerprint(s string) ([]byte, error) {
	h := strings.TrimSpace(s)
	if len(h) >= 7 && strings.EqualFold(h[:7], "sha256:") {
		h = h[7:]
	}
	h = strings.ReplaceAll(h, ":", "")
	raw, err := hex.DecodeString(h)
	if err != nil || len(raw) != sha256.Size {
		return nil, errors.New("tls: cert_fingerprint must be a SHA-256 fingerprint (64 hex characters)")
	}
	return raw, nil
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
		var certErr *tls.CertificateVerificationError
		switch {
		case errors.As(err, &certErr), errors.Is(err, errPinMismatch):
			return fmt.Errorf("TLS certificate for %s is not trusted (%w) — pin it with cert_fingerprint or trust its CA with ca_file", c.endpoint, err)
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
