package store

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/eSlider/self-ca/internal/model"
)

type Filesystem struct {
	mu  sync.RWMutex
	dir string
}

func NewFilesystem(dir string) *Filesystem {
	return &Filesystem{dir: dir}
}

func (f *Filesystem) CreateCA(_ context.Context, ca model.CA) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	caDir := f.caDir(ca.ID)
	if _, err := os.Stat(caDir); err == nil {
		return ErrAlreadyExists
	}
	if err := os.MkdirAll(filepath.Join(caDir, "certs"), 0o755); err != nil {
		return err
	}
	return f.writeCA(ca)
}

func (f *Filesystem) GetCA(ctx context.Context, id string) (model.CA, error) {
	ca, err := f.GetCAWithKey(ctx, id)
	if err != nil {
		return model.CA{}, err
	}
	return sanitizeCA(ca), nil
}

func (f *Filesystem) GetCAWithKey(_ context.Context, id string) (model.CA, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	caDir := f.caDir(id)
	if _, err := os.Stat(caDir); os.IsNotExist(err) {
		return model.CA{}, ErrNotFound
	}

	meta, err := readJSON[model.CA](filepath.Join(caDir, "meta.json"))
	if err != nil {
		return model.CA{}, err
	}
	certPEM, err := os.ReadFile(filepath.Join(caDir, "ca.crt"))
	if err != nil {
		return model.CA{}, err
	}
	keyPEM, err := os.ReadFile(filepath.Join(caDir, "ca.key"))
	if err != nil {
		return model.CA{}, err
	}
	meta.CertPEM = string(certPEM)
	meta.KeyPEM = string(keyPEM)
	return meta, nil
}

func (f *Filesystem) ListCAs(_ context.Context) ([]model.CA, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	entries, err := os.ReadDir(f.casRoot())
	if err != nil {
		if os.IsNotExist(err) {
			return []model.CA{}, nil
		}
		return nil, err
	}

	out := make([]model.CA, 0)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		meta, err := readJSON[model.CA](filepath.Join(f.caDir(entry.Name()), "meta.json"))
		if err != nil {
			return nil, err
		}
		out = append(out, sanitizeCA(meta))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func (f *Filesystem) UpdateCA(_ context.Context, id string, patch model.CAUpdate) (model.CA, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	caDir := f.caDir(id)
	if _, err := os.Stat(caDir); os.IsNotExist(err) {
		return model.CA{}, ErrNotFound
	}

	meta, err := readJSON[model.CA](filepath.Join(caDir, "meta.json"))
	if err != nil {
		return model.CA{}, err
	}
	if patch.CommonName != nil {
		meta.CommonName = *patch.CommonName
	}
	if patch.Organization != nil {
		meta.Organization = *patch.Organization
	}
	if patch.Country != nil {
		meta.Country = *patch.Country
	}
	if patch.Province != nil {
		meta.Province = *patch.Province
	}
	if patch.Locality != nil {
		meta.Locality = *patch.Locality
	}
	if err := writeJSON(filepath.Join(caDir, "meta.json"), meta); err != nil {
		return model.CA{}, err
	}
	return sanitizeCA(meta), nil
}

func (f *Filesystem) DeleteCA(_ context.Context, id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	caDir := f.caDir(id)
	if _, err := os.Stat(caDir); os.IsNotExist(err) {
		return ErrNotFound
	}
	return os.RemoveAll(caDir)
}

func (f *Filesystem) CreateCert(_ context.Context, caID string, cert model.LeafCert) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	caDir := f.caDir(caID)
	if _, err := os.Stat(caDir); os.IsNotExist(err) {
		return ErrNotFound
	}
	certDir := f.certDir(caID, cert.ID)
	if _, err := os.Stat(certDir); err == nil {
		return ErrAlreadyExists
	}
	if err := os.MkdirAll(certDir, 0o755); err != nil {
		return err
	}
	return f.writeCert(caID, cert)
}

func (f *Filesystem) GetCert(_ context.Context, caID, id string) (model.LeafCert, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	certDir := f.certDir(caID, id)
	if _, err := os.Stat(certDir); os.IsNotExist(err) {
		return model.LeafCert{}, ErrNotFound
	}

	meta, err := readJSON[model.LeafCert](filepath.Join(certDir, "meta.json"))
	if err != nil {
		return model.LeafCert{}, err
	}
	certPEM, err := os.ReadFile(filepath.Join(certDir, "cert.pem"))
	if err != nil {
		return model.LeafCert{}, err
	}
	keyPEM, err := os.ReadFile(filepath.Join(certDir, "key.pem"))
	if err != nil {
		return model.LeafCert{}, err
	}
	meta.CertPEM = string(certPEM)
	meta.KeyPEM = string(keyPEM)
	return meta, nil
}

func (f *Filesystem) ListCerts(_ context.Context, caID string) ([]model.LeafCert, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	certsDir := filepath.Join(f.caDir(caID), "certs")
	entries, err := os.ReadDir(certsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	out := make([]model.LeafCert, 0)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		meta, err := readJSON[model.LeafCert](filepath.Join(certsDir, entry.Name(), "meta.json"))
		if err != nil {
			return nil, err
		}
		out = append(out, sanitizeCert(meta))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func (f *Filesystem) UpdateCert(_ context.Context, caID, id string, cert model.LeafCert) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	certDir := f.certDir(caID, id)
	if _, err := os.Stat(certDir); os.IsNotExist(err) {
		return ErrNotFound
	}
	return f.writeCert(caID, cert)
}

func (f *Filesystem) DeleteCert(_ context.Context, caID, id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	certDir := f.certDir(caID, id)
	if _, err := os.Stat(certDir); os.IsNotExist(err) {
		return ErrNotFound
	}
	return os.RemoveAll(certDir)
}

func (f *Filesystem) casRoot() string {
	return filepath.Join(f.dir, "cas")
}

func (f *Filesystem) caDir(id string) string {
	return filepath.Join(f.casRoot(), id)
}

func (f *Filesystem) certDir(caID, certID string) string {
	return filepath.Join(f.caDir(caID), "certs", certID)
}

func (f *Filesystem) writeCA(ca model.CA) error {
	caDir := f.caDir(ca.ID)
	if err := os.MkdirAll(caDir, 0o755); err != nil {
		return err
	}
	meta := ca
	meta.CertPEM = ""
	meta.KeyPEM = ""
	if err := writeJSON(filepath.Join(caDir, "meta.json"), meta); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(caDir, "ca.crt"), []byte(ca.CertPEM), 0o644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(caDir, "ca.key"), []byte(ca.KeyPEM), 0o600)
}

func (f *Filesystem) writeCert(_ string, cert model.LeafCert) error {
	certDir := f.certDir(cert.CAID, cert.ID)
	meta := cert
	meta.CertPEM = ""
	meta.KeyPEM = ""
	if err := writeJSON(filepath.Join(certDir, "meta.json"), meta); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(certDir, "cert.pem"), []byte(cert.CertPEM), 0o644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(certDir, "key.pem"), []byte(cert.KeyPEM), 0o600)
}

func readJSON[T any](path string) (T, error) {
	var v T
	b, err := os.ReadFile(path)
	if err != nil {
		return v, err
	}
	if err := json.Unmarshal(b, &v); err != nil {
		return v, err
	}
	return v, nil
}

func writeJSON(path string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o600)
}

func NewFromConfig(dataDir string) Store {
	if dataDir == "" {
		return NewMemory()
	}
	return NewFilesystem(dataDir)
}
