package ca

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"time"
)

type CAOptions struct {
	CommonName   string
	Organization string
	Country      string
	Province     string
	Locality     string
	ValidYears   int
}

type LeafOptions struct {
	CommonName  string
	DNSNames    []string
	IPAddresses []string
	ValidYears  int
}

type GeneratedCA struct {
	CertPEM []byte
	KeyPEM  []byte
	Cert    *x509.Certificate
	Key     *ecdsa.PrivateKey
}

type GeneratedLeaf struct {
	CertPEM []byte
	KeyPEM  []byte
	Cert    *x509.Certificate
	Key     *ecdsa.PrivateKey
}

func GenerateCA(opts CAOptions) (*GeneratedCA, error) {
	if opts.CommonName == "" {
		return nil, fmt.Errorf("common_name is required")
	}
	if opts.ValidYears <= 0 {
		opts.ValidYears = 10
	}

	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generate CA key: %w", err)
	}

	template := &x509.Certificate{
		SerialNumber: serialNumber(),
		Subject: pkix.Name{
			Organization: nonEmptySlice(opts.Organization),
			Country:      nonEmptySlice(opts.Country),
			Province:     nonEmptySlice(opts.Province),
			Locality:     nonEmptySlice(opts.Locality),
			CommonName:   opts.CommonName,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(opts.ValidYears, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &privKey.PublicKey, privKey)
	if err != nil {
		return nil, fmt.Errorf("create CA certificate: %w", err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyPEM, err := marshalPrivateKey(privKey)
	if err != nil {
		return nil, err
	}

	parsed, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, fmt.Errorf("parse CA certificate: %w", err)
	}

	return &GeneratedCA{
		CertPEM: certPEM,
		KeyPEM:  keyPEM,
		Cert:    parsed,
		Key:     privKey,
	}, nil
}

func IssueLeaf(caCert *x509.Certificate, caKey *ecdsa.PrivateKey, opts LeafOptions) (*GeneratedLeaf, error) {
	if caCert == nil || caKey == nil {
		return nil, fmt.Errorf("CA certificate and key are required")
	}
	if opts.CommonName == "" {
		return nil, fmt.Errorf("common_name is required")
	}
	if opts.ValidYears <= 0 {
		opts.ValidYears = 1
	}

	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generate leaf key: %w", err)
	}

	ips := make([]net.IP, 0, len(opts.IPAddresses))
	for _, raw := range opts.IPAddresses {
		if ip := net.ParseIP(raw); ip != nil {
			ips = append(ips, ip)
		}
	}

	template := &x509.Certificate{
		SerialNumber: serialNumber(),
		Subject: pkix.Name{
			CommonName: opts.CommonName,
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().AddDate(opts.ValidYears, 0, 0),
		KeyUsage:    x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:    append([]string(nil), opts.DNSNames...),
		IPAddresses: ips,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, caCert, &privKey.PublicKey, caKey)
	if err != nil {
		return nil, fmt.Errorf("create leaf certificate: %w", err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyPEM, err := marshalPrivateKey(privKey)
	if err != nil {
		return nil, err
	}

	parsed, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, fmt.Errorf("parse leaf certificate: %w", err)
	}

	return &GeneratedLeaf{
		CertPEM: certPEM,
		KeyPEM:  keyPEM,
		Cert:    parsed,
		Key:     privKey,
	}, nil
}

func ParseCA(certPEM, keyPEM []byte) (*x509.Certificate, *ecdsa.PrivateKey, error) {
	certBlock, _ := pem.Decode(certPEM)
	if certBlock == nil || certBlock.Type != "CERTIFICATE" {
		return nil, nil, fmt.Errorf("invalid certificate PEM")
	}

	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("parse certificate: %w", err)
	}

	keyBlock, _ := pem.Decode(keyPEM)
	if keyBlock == nil {
		return nil, nil, fmt.Errorf("invalid private key PEM")
	}

	key, err := x509.ParseECPrivateKey(keyBlock.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("parse private key: %w", err)
	}

	return cert, key, nil
}

func marshalPrivateKey(key *ecdsa.PrivateKey) ([]byte, error) {
	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("marshal private key: %w", err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER}), nil
}

func serialNumber() *big.Int {
	n, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return big.NewInt(time.Now().UnixNano())
	}
	return n
}

func nonEmptySlice(value string) []string {
	if value == "" {
		return nil
	}
	return []string{value}
}
