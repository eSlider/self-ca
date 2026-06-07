package main

import (
	"crypto/tls"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/eSlider/self-ca/internal/config"
)

func TestHTTPServer(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "cert-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfgPath := filepath.Join(tempDir, "config.yml")
	cfgContent := `setup:
  output:
    ca_cert: ca.crt
    server_cert: server.crt
    server_key: server.key
`
	if err := os.WriteFile(cfgPath, []byte(cfgContent), 0o644); err != nil {
		t.Fatal(err)
	}

	originalWd, _ := os.Getwd()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to chdir: %v", err)
	}
	defer os.Chdir(originalWd)

	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("Load config: %v", err)
	}
	if err := generateCertsToDisk(cfg); err != nil {
		t.Fatalf("Failed to generate certs: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, HTTPS"))
	})

	server := httptest.NewUnstartedServer(handler)

	cert, err := tls.LoadX509KeyPair("server.crt", "server.key")
	if err != nil {
		t.Fatalf("Failed to load server key pair: %v", err)
	}

	server.TLS = &tls.Config{Certificates: []tls.Certificate{cert}}
	server.StartTLS()
	defer server.Close()

	client := server.Client()

	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("Failed to make HTTPS request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, but got %d", http.StatusOK, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	expectedBody := "Hello, HTTPS"
	if string(body) != expectedBody {
		t.Errorf("Expected body '%s', but got '%s'", expectedBody, string(body))
	}
}

func TestGenerateCertsFromConfig(t *testing.T) {
	tempDir := t.TempDir()
	cfg := config.Default()
	cfg.Setup.Output.CACert = filepath.Join(tempDir, "ca.crt")
	cfg.Setup.Output.ServerCert = filepath.Join(tempDir, "server.crt")
	cfg.Setup.Output.ServerKey = filepath.Join(tempDir, "server.key")

	if err := generateCertsToDisk(cfg); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{cfg.Setup.Output.CACert, cfg.Setup.Output.ServerCert, cfg.Setup.Output.ServerKey} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("missing %s: %v", path, err)
		}
	}
}
