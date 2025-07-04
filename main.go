package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"flag"
	"log"
	"math/big"
	"net"
	"net/http"
	"os"
	"time"
)

// Go program to set up a self-signed HTTPS server with ECDSA certificates.
func main() {
	setup := flag.Bool("setup", false, "Generate CA and server certificates")
	flag.Parse()

	if *setup {
		log.Println("Setting up new CA and server certificates...")
		generateCerts()
		log.Println("Setup complete. Certificates created: ca.crt, server.crt, server.key")
		return
	}

	log.Println("Starting HTTPS server on :8443...")
	// Check if certs exist before starting server
	if _, err := os.Stat("server.crt"); os.IsNotExist(err) {
		log.Fatalf("server.crt not found. Please run with -setup flag to generate certificates.")
	}
	if _, err := os.Stat("server.key"); os.IsNotExist(err) {
		log.Fatalf("server.key not found. Please run with -setup flag to generate certificates.")
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, HTTPS"))
	})

	if err := http.ListenAndServeTLS(":8443", "server.crt", "server.key", nil); err != nil {
		log.Fatalf("ListenAndServeTLS error: %v", err)
	}
}

func generateCerts() {
	// Generate CA private key
	caPrivKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatalf("Failed to generate CA private key: %v", err)
	}

	// Create CA certificate template
	caTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Produktor"},
			Country:      []string{"UA"},
			Province:     []string{"Ukraine"},
			Locality:     []string{"Dnepr"},
			CommonName:   "localhost CA",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0), // 10 year validity
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	// Create CA certificate
	caBytes, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		log.Fatalf("Failed to create CA certificate: %v", err)
	}

	// Encode CA certificate to PEM
	caPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})
	if err := os.WriteFile("ca.crt", caPEM, 0644); err != nil {
		log.Fatalf("Failed to write CA certificate to file: %v", err)
	}

	// Generate server private key
	serverPrivKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatalf("Failed to generate server private key: %v", err)
	}

	// Encode server private key to PEM
	serverPrivKeyBytes, err := x509.MarshalECPrivateKey(serverPrivKey)
	if err != nil {
		log.Fatalf("Unable to marshal ECDSA private key: %v", err)
	}
	serverPrivKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: serverPrivKeyBytes,
	})
	if err := os.WriteFile("server.key", serverPrivKeyPEM, 0600); err != nil {
		log.Fatalf("Failed to write server key to file: %v", err)
	}

	// Create server certificate template
	serverTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject: pkix.Name{
			Organization: []string{"Produktor"},
			Country:      []string{"UA"},
			Province:     []string{"Ukraine"},
			Locality:     []string{"Dnepr"},
			CommonName:   "localhost",
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().AddDate(1, 0, 0),
		KeyUsage:    x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:    []string{"localhost"},
		IPAddresses: []net.IP{net.ParseIP("127.0.0.1")},
	}

	// Create server certificate, signed by our CA
	serverCertBytes, err := x509.CreateCertificate(rand.Reader, serverTemplate, caTemplate, &serverPrivKey.PublicKey, caPrivKey)
	if err != nil {
		log.Fatalf("Failed to create server certificate: %v", err)
	}

	// Encode server certificate to PEM
	serverCertPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: serverCertBytes,
	})
	if err := os.WriteFile("server.crt", serverCertPEM, 0644); err != nil {
		log.Fatalf("Failed to write server certificate to file: %v", err)
	}
}

