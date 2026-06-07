package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/eSlider/self-ca/internal/api"
	"github.com/eSlider/self-ca/internal/ca"
	"github.com/eSlider/self-ca/internal/config"
	"github.com/eSlider/self-ca/internal/store"
	"github.com/eSlider/self-ca/internal/web"
)

func main() {
	var (
		configPath  = flag.String("config", "config.yml", "Path to config.yml")
		setup       = flag.Bool("setup", false, "Generate CA and server certificates from config")
		apiOverride = flag.String("api", "", "Override server.api_addr (empty disables API/UI HTTP)")
		tlsOverride = flag.String("tls", "", "Override server.tls_addr (empty disables HTTPS)")
	)
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Load config: %v", err)
	}
	if *apiOverride != "" {
		cfg.Server.APIAddr = *apiOverride
	}
	if *tlsOverride != "" {
		cfg.Server.TLSAddr = *tlsOverride
	}

	if *setup {
		log.Println("Setting up CA and server certificates from config...")
		if err := generateCertsToDisk(cfg); err != nil {
			log.Fatalf("Setup failed: %v", err)
		}
		log.Printf("Setup complete: %s, %s, %s",
			cfg.Setup.Output.CACert, cfg.Setup.Output.ServerCert, cfg.Setup.Output.ServerKey)
		return
	}

	dataDir := cfg.Data.Dir
	st := store.NewFilesystem(dataDir)
	handler := newHandler(st)

	httpAddr := cfg.Server.APIAddr
	tlsAddr := cfg.Server.TLSAddr
	tlsReady := tlsCertsPresent(cfg.Server.TLSCert, cfg.Server.TLSKey)

	if httpAddr != "" {
		go serveHTTP(httpAddr, handler, dataDir)
	}

	if tlsAddr != "" && tlsReady {
		log.Printf("Starting HTTPS on %s (UI + API)...", tlsAddr)
		serveTLS(tlsAddr, handler, cfg.Server.TLSCert, cfg.Server.TLSKey)
		return
	}

	if tlsAddr != "" && !tlsReady {
		log.Printf("WARN: TLS certs missing (%s, %s) — HTTP-only mode (SEC-1). Run with -setup to enable HTTPS.",
			cfg.Server.TLSCert, cfg.Server.TLSKey)
	}

	if httpAddr == "" {
		log.Fatal("No listeners configured: set server.api_addr and/or provide TLS certs for server.tls_addr")
	}

	select {}
}

func newHandler(st store.Store) http.Handler {
	mux := http.NewServeMux()
	api.NewHandler(st).Register(mux)
	mux.Handle("/", web.Handler())
	return mux
}

func serveHTTP(addr string, handler http.Handler, dataDir string) {
	srv := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	log.Printf("Starting HTTP on %s (UI + API, data: %s)...", addr, dataDir)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("HTTP server error: %v", err)
	}
}

func serveTLS(addr string, handler http.Handler, certFile, keyFile string) {
	srv := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	if err := srv.ListenAndServeTLS(certFile, keyFile); err != nil {
		log.Fatalf("HTTPS server error: %v", err)
	}
}

func tlsCertsPresent(certFile, keyFile string) bool {
	for _, path := range []string{certFile, keyFile} {
		if _, err := os.Stat(path); err != nil {
			return false
		}
	}
	return true
}

func generateCertsToDisk(cfg config.Config) error {
	generatedCA, err := ca.GenerateCA(cfg.Setup.CAOptions())
	if err != nil {
		return err
	}
	if err := os.WriteFile(cfg.Setup.Output.CACert, generatedCA.CertPEM, 0o644); err != nil { // #nosec G306 -- public CA certificate
		return err
	}

	generatedLeaf, err := ca.IssueLeaf(generatedCA.Cert, generatedCA.Key, cfg.Setup.LeafOptions())
	if err != nil {
		return err
	}
	if err := os.WriteFile(cfg.Setup.Output.ServerCert, generatedLeaf.CertPEM, 0o644); err != nil { // #nosec G306 -- public server certificate
		return err
	}
	if err := os.WriteFile(cfg.Setup.Output.ServerKey, generatedLeaf.KeyPEM, 0o600); err != nil {
		return err
	}
	return nil
}
