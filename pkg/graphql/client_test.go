package graphql

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const testSecret = "s3cret"

func okHandler(seen *[]string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*seen = append(*seen, r.Header.Get("x-hasura-admin-secret"))
		_, _ = w.Write([]byte(`{"data":{}}`))
	})
}

func do(t *testing.T, endpoint string, opts TLSOptions) error {
	t.Helper()
	c, err := NewClient(endpoint, testSecret, opts, 5*time.Second)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	var out map[string]any
	return c.Do(context.Background(), `{ x }`, &out)
}

func fingerprint(srv *httptest.Server) string {
	sum := sha256.Sum256(srv.Certificate().Raw)
	return hex.EncodeToString(sum[:])
}

func TestTLSVerification(t *testing.T) {
	var seen []string
	srv := httptest.NewTLSServer(okHandler(&seen))
	defer srv.Close()
	other := httptest.NewTLSServer(okHandler(&seen))
	defer other.Close()

	caFile := filepath.Join(t.TempDir(), "ca.pem")
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: srv.Certificate().Raw})
	if err := os.WriteFile(caFile, pemBytes, 0o600); err != nil {
		t.Fatal(err)
	}

	colonFP := strings.ToUpper(fingerprint(srv))
	var parts []string
	for i := 0; i < len(colonFP); i += 2 {
		parts = append(parts, colonFP[i:i+2])
	}

	tests := []struct {
		name    string
		opts    TLSOptions
		wantErr bool
	}{
		{"untrusted cert is refused by default", TLSOptions{}, true},
		{"ca_file trusts the cert", TLSOptions{CAFile: caFile}, false},
		{"matching fingerprint", TLSOptions{CertFingerprint: fingerprint(srv)}, false},
		{"colon-separated fingerprint", TLSOptions{CertFingerprint: "SHA256:" + strings.Join(parts, ":")}, false},
		{"wrong fingerprint is refused", TLSOptions{CertFingerprint: strings.Repeat("00", 32)}, true},
		{"insecure_skip_verify opts out", TLSOptions{InsecureSkipVerify: true}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seen = nil
			err := do(t, srv.URL, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				if len(seen) != 0 {
					t.Fatalf("admin secret sent despite failed verification")
				}
				if !strings.Contains(err.Error(), "not trusted") {
					t.Fatalf("unhelpful error: %v", err)
				}
			}
		})
	}
}

func TestTLSOptionsRejected(t *testing.T) {
	for _, opts := range []TLSOptions{
		{CertFingerprint: "abcd"},
		{CAFile: filepath.Join(t.TempDir(), "missing.pem")},
		{CertFingerprint: strings.Repeat("00", 32), InsecureSkipVerify: true},
	} {
		if _, err := NewClient("https://127.0.0.1:1/", testSecret, opts, time.Second); err == nil {
			t.Errorf("NewClient(%+v) succeeded, want error", opts)
		}
	}
}

func TestRedirects(t *testing.T) {
	var otherSeen, sameSeen []string
	other := httptest.NewServer(okHandler(&otherSeen))
	defer other.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/cross", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, other.URL+"/graphql/", http.StatusTemporaryRedirect)
	})
	mux.HandleFunc("/same", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/graphql/", http.StatusTemporaryRedirect)
	})
	mux.Handle("/graphql/", okHandler(&sameSeen))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	if err := do(t, srv.URL+"/cross", TLSOptions{}); err == nil {
		t.Fatal("cross-origin redirect was followed")
	}
	if len(otherSeen) != 0 {
		t.Fatalf("admin secret forwarded to another origin: %q", otherSeen)
	}

	if err := do(t, srv.URL+"/same", TLSOptions{}); err != nil {
		t.Fatalf("same-origin redirect: %v", err)
	}
	if len(sameSeen) != 1 || sameSeen[0] != testSecret {
		t.Fatalf("same-origin redirect lost the secret: %q", sameSeen)
	}
}
