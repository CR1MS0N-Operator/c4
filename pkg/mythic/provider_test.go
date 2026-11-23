package mythic

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/CR1MS0N-Operator/c4/pkg/config"
	"github.com/CR1MS0N-Operator/c4/pkg/graphql"
	"github.com/CR1MS0N-Operator/c4/pkg/provider"
)

func operatorHandler(hits *atomic.Int32) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		_, _ = w.Write([]byte(`{"data":{"operator":[{"id":1,"username":"op","current_operation_id":1}]}}`))
	})
}

func port(t *testing.T, srv *httptest.Server) int {
	t.Helper()
	_, p, err := net.SplitHostPort(srv.Listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	n, err := strconv.Atoi(p)
	if err != nil {
		t.Fatal(err)
	}
	return n
}

// closedPort returns a loopback port nothing is listening on.
func closedPort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	p := l.Addr().(*net.TCPAddr).Port
	l.Close()
	return p
}

func TestConnectRefusesUntrustedCert(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewTLSServer(operatorHandler(&hits))
	defer srv.Close()

	cfg := config.C2Config{Host: "127.0.0.1", APIPort: port(t, srv), HasuraSecret: "s", SSL: true}
	if err := NewMythicProvider("mythic", cfg).Connect(context.Background()); err == nil {
		t.Fatal("connected to a server with an untrusted certificate")
	}
	if hits.Load() != 0 {
		t.Fatal("request reached the untrusted server")
	}
}

func TestConnectServerPortFallback(t *testing.T) {
	var plainHits, tlsHits atomic.Int32
	plain := httptest.NewServer(operatorHandler(&plainHits))
	defer plain.Close()
	tlsSrv := httptest.NewTLSServer(operatorHandler(&tlsHits))
	defer tlsSrv.Close()

	base := config.C2Config{Host: "127.0.0.1", APIPort: closedPort(t), HasuraSecret: "s", SSL: true, InsecureSkipVerify: true}

	t.Run("off by default", func(t *testing.T) {
		cfg := base
		cfg.ServerPort = port(t, plain)
		if err := NewMythicProvider("mythic", cfg).Connect(context.Background()); err == nil {
			t.Fatal("connected without server_port_fallback")
		}
		if plainHits.Load() != 0 {
			t.Fatal("fell back to server_port without server_port_fallback")
		}
	})

	t.Run("never downgrades to http", func(t *testing.T) {
		cfg := base
		cfg.ServerPort = port(t, plain)
		cfg.ServerPortFallback = true
		if err := NewMythicProvider("mythic", cfg).Connect(context.Background()); err == nil {
			t.Fatal("connected to a plain-HTTP backend with ssl = true")
		}
		if plainHits.Load() != 0 {
			t.Fatal("admin secret sent over plain HTTP")
		}
	})

	t.Run("same scheme when enabled", func(t *testing.T) {
		cfg := base
		cfg.ServerPort = port(t, tlsSrv)
		cfg.ServerPortFallback = true
		if err := NewMythicProvider("mythic", cfg).Connect(context.Background()); err != nil {
			t.Fatalf("fallback: %v", err)
		}
		if tlsHits.Load() == 0 {
			t.Fatal("fallback endpoint not used")
		}
	})
}

func TestStartListenerSendsNameAsVariable(t *testing.T) {
	const name = `a\"}) { id } x: __typename #`

	type gqlRequest struct {
		Query     string         `json:"query"`
		Variables map[string]any `json:"variables"`
	}
	var mu sync.Mutex
	var reqs []gqlRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req gqlRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode request: %v", err)
		}
		mu.Lock()
		reqs = append(reqs, req)
		mu.Unlock()
		if strings.Contains(req.Query, "insert_c2profile_one") {
			_, _ = w.Write([]byte(`{"data":{"insert_c2profile_one":{"id":7,"name":"n","running":false}}}`))
			return
		}
		_, _ = w.Write([]byte(`{"data":{"c2profile":[]}}`))
	}))
	defer srv.Close()

	client, err := graphql.NewClient(srv.URL+"/graphql/", "s", graphql.TLSOptions{}, 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	p := &MythicProvider{name: "mythic", client: client}
	if _, err := p.StartListener(context.Background(), provider.Listener{Name: name}); err != nil {
		t.Fatalf("StartListener: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(reqs) != 2 {
		t.Fatalf("got %d requests, want 2", len(reqs))
	}
	wantQueries := []string{
		`query($name: String!) { c2profile(where: {name: {_eq: $name}}) { id name running } }`,
		`mutation($name: String!, $description: String!) { insert_c2profile_one(object: {name: $name, description: $description}) { id name running } }`,
	}
	for i, req := range reqs {
		if req.Query != wantQueries[i] {
			t.Errorf("request %d query is not the constant document: %s", i, req.Query)
		}
		if req.Variables["name"] != name {
			t.Errorf("variables.name = %v, want %q", req.Variables["name"], name)
		}
	}
	if want := "Managed by C4 — " + name; reqs[1].Variables["description"] != want {
		t.Errorf("variables.description = %v, want %q", reqs[1].Variables["description"], want)
	}
}
