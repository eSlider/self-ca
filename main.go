package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/eSlider/self-ca/internal/api"
	"github.com/eSlider/self-ca/internal/ca"
	"github.com/eSlider/self-ca/internal/config"
	"github.com/eSlider/self-ca/internal/store"
)

func main() {
	var (
		configPath  = flag.String("config", "config.yml", "Path to config.yml")
		setup       = flag.Bool("setup", false, "Generate CA and server certificates from config")
		apiOverride = flag.String("api", "", "Override server.api_addr (empty disables API)")
		tlsOverride = flag.String("tls", "", "Override server.tls_addr")
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

	if cfg.Server.APIAddr != "" {
		go func() {
			mux := http.NewServeMux()
			api.NewHandler(st).Register(mux)
			log.Printf("Starting HTTP API on %s (data: %s)...", cfg.Server.APIAddr, dataDir)
			if err := http.ListenAndServe(cfg.Server.APIAddr, mux); err != nil {
				log.Fatalf("API server error: %v", err)
			}
		}()
	}

	log.Printf("Starting HTTPS server on %s...", cfg.Server.TLSAddr)
	if _, err := os.Stat(cfg.Server.TLSCert); os.IsNotExist(err) {
		log.Fatalf("%s not found. Run with -setup to generate certificates.", cfg.Server.TLSCert)
	}
	if _, err := os.Stat(cfg.Server.TLSKey); os.IsNotExist(err) {
		log.Fatalf("%s not found. Run with -setup to generate certificates.", cfg.Server.TLSKey)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Hello, HTTPS"))
	})

	if err := http.ListenAndServeTLS(cfg.Server.TLSAddr, cfg.Server.TLSCert, cfg.Server.TLSKey, nil); err != nil {
		log.Fatalf("ListenAndServeTLS error: %v", err)
	}
}

func generateCertsToDisk(cfg config.Config) error {
	generatedCA, err := ca.GenerateCA(cfg.Setup.CAOptions())
	if err != nil {
		return err
	}
	if err := os.WriteFile(cfg.Setup.Output.CACert, generatedCA.CertPEM, 0o644); err != nil {
		return err
	}

	generatedLeaf, err := ca.IssueLeaf(generatedCA.Cert, generatedCA.Key, cfg.Setup.LeafOptions())
	if err != nil {
		return err
	}
	if err := os.WriteFile(cfg.Setup.Output.ServerCert, generatedLeaf.CertPEM, 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(cfg.Setup.Output.ServerKey, generatedLeaf.KeyPEM, 0o600); err != nil {
		return err
	}
	return nil
}
