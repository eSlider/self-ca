package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/eSlider/self-ca/internal/config"
)

func TestLoad_FromYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")
	const content = `
server:
  api_addr: ":9090"
  tls_addr: ":9443"
  tls_cert: custom.crt
  tls_key: custom.key
data:
  dir: ./my-data
setup:
  ca:
    common_name: Test CA
    organization: Acme
    country: US
    valid_years: 5
  server:
    common_name: dev.local
    dns_names:
      - dev.local
      - localhost
    ip_addresses:
      - 127.0.0.1
    valid_years: 2
  output:
    ca_cert: my-ca.crt
    server_cert: my-server.crt
    server_key: my-server.key
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.Server.APIAddr != ":9090" {
		t.Fatalf("api_addr = %q", cfg.Server.APIAddr)
	}
	if cfg.Server.TLSAddr != ":9443" {
		t.Fatalf("tls_addr = %q", cfg.Server.TLSAddr)
	}
	if cfg.Data.Dir != "./my-data" {
		t.Fatalf("data.dir = %q", cfg.Data.Dir)
	}
	if cfg.Setup.CA.CommonName != "Test CA" {
		t.Fatalf("setup.ca.common_name = %q", cfg.Setup.CA.CommonName)
	}
	if cfg.Setup.Server.DNSNames[0] != "dev.local" {
		t.Fatalf("dns_names = %v", cfg.Setup.Server.DNSNames)
	}
	if cfg.Setup.Output.CACert != "my-ca.crt" {
		t.Fatalf("output.ca_cert = %q", cfg.Setup.Output.CACert)
	}
}

func TestLoad_EnvOverridesYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")
	if err := os.WriteFile(path, []byte(`server:
  api_addr: ":8080"
`), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("SERVER_APIADDR", ":3000")

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Server.APIAddr != ":3000" {
		t.Fatalf("api_addr = %q, want :3000 from env", cfg.Server.APIAddr)
	}
}

func TestLoad_MissingFileUsesDefaults(t *testing.T) {
	cfg, err := config.Load(filepath.Join(t.TempDir(), "missing.yml"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Server.APIAddr != ":8080" {
		t.Fatalf("default api_addr = %q", cfg.Server.APIAddr)
	}
	if cfg.Setup.CA.CommonName != "localhost CA" {
		t.Fatalf("default CA CN = %q", cfg.Setup.CA.CommonName)
	}
}

func TestSetupCAOptions(t *testing.T) {
	cfg := config.Default()
	opts := cfg.Setup.CAOptions()
	if opts.CommonName != "localhost CA" || opts.ValidYears != 10 {
		t.Fatalf("CAOptions = %+v", opts)
	}
}

func TestSetupLeafOptions(t *testing.T) {
	cfg := config.Default()
	opts := cfg.Setup.LeafOptions()
	if opts.CommonName != "localhost" || len(opts.DNSNames) != 1 {
		t.Fatalf("LeafOptions = %+v", opts)
	}
}
