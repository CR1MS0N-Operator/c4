package mythic

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync/atomic"
	"testing"

	"github.com/CR1MS0N-Operator/c4/pkg/config"
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
