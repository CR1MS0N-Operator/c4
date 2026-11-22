package cmd

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/CR1MS0N-Operator/c4/pkg/provider"
)

// captureStdout runs fn and returns everything it wrote to os.Stdout.
func captureStdout(t *testing.T, fn func()) []byte {
	t.Helper()
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w
	defer func() { os.Stdout = orig }()

	fn()
	w.Close()
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	return out
}

func TestPrintCallbackTableStripsEscapes(t *testing.T) {
	out := captureStdout(t, func() {
		printCallbackTable([]provider.Callback{{
			ID:    "1",
			Agent: "apollo\x1b[2J",
			Host:  "host\x1b[1A\x1b[2K",
			User:  "a\x1b]52;c;ZWNobyBoaQ==\x07b",
		}})
	})
	if bytes.IndexByte(out, 0x1b) != -1 {
		t.Fatalf("callback table output contains ESC (0x1b): %q", out)
	}
	if bytes.IndexByte(out, 0x07) != -1 {
		t.Fatalf("callback table output contains BEL (0x07): %q", out)
	}
}

func TestPrintPayloadAndListenerTablesStripEscapes(t *testing.T) {
	out := captureStdout(t, func() {
		printPayloadTable([]provider.Payload{{ID: "1", Name: "p\x1b[2J", Type: "t\x1b[0m", Format: "f\x9b2J", FilePath: "u\x1b]0;x\x07"}})
		printListenerTable([]provider.Listener{{ID: "1", Name: "l\x1b[2J", Type: "http\x1b[0m"}})
		printHealthTable(rootLogger, "n\x1b[2J", "Mythic", "Running", "h\x1b[1A", "v", "msg\x1b]52;c;AA==\x07")
	})
	if bytes.IndexByte(out, 0x1b) != -1 || bytes.IndexByte(out, 0x07) != -1 {
		t.Fatalf("table output contains control bytes: %q", out)
	}
}

func TestSanitizeFieldStripsEscapes(t *testing.T) {
	// OSC 52 clipboard-write payload embedded in a self-reported user field.
	in := "a\x1b]52;c;ZWNobyBoaQ==\x07b"
	out := sanitizeField(in)
	if strings.ContainsRune(out, 0x1b) {
		t.Fatalf("output still contains ESC (0x1b): %q", out)
	}
	if bytes.ContainsAny([]byte(out), "\x07") {
		t.Fatalf("output still contains BEL (0x07): %q", out)
	}
	if out != "a?]52;c;ZWNobyBoaQ==?b" {
		t.Fatalf("unexpected sanitized output: %q", out)
	}
}

func TestSanitizeFieldPreservesPrintable(t *testing.T) {
	cases := []string{"host-01", "DESKTOP\\user", "αβγ", "10.0.0.5:8080", ""}
	for _, c := range cases {
		if got := sanitizeField(c); got != c {
			t.Errorf("printable input %q altered to %q", c, got)
		}
	}
}

func TestSanitizeFieldReplacesControlChars(t *testing.T) {
	for _, r := range []rune{0x00, 0x1b, 0x7f, 0x9b} {
		if got := sanitizeField(string(r)); got != "?" {
			t.Errorf("control rune %#x not replaced: %q", r, got)
		}
	}
}

func TestSanitizeFieldNeutralizesRawC1Bytes(t *testing.T) {
	// A raw 0x9b byte (8-bit CSI) is invalid UTF-8 and must not survive.
	out := sanitizeField("f\x9b2J")
	if bytes.IndexByte([]byte(out), 0x9b) != -1 {
		t.Fatalf("raw C1 byte survived: %q", out)
	}
}
