package store

import (
	"context"
	"sort"
	"sync"

	"github.com/eSlider/self-ca/internal/model"
)

type Memory struct {
	mu    sync.RWMutex
	cas   map[string]model.CA
	certs map[string]map[string]model.LeafCert
}

func NewMemory() *Memory {
	return &Memory{
		cas:   make(map[string]model.CA),
		certs: make(map[string]map[string]model.LeafCert),
	}
}

func (m *Memory) CreateCA(_ context.Context, ca model.CA) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.cas[ca.ID]; ok {
		return ErrAlreadyExists
	}
	m.cas[ca.ID] = ca
	m.certs[ca.ID] = make(map[string]model.LeafCert)
	return nil
}

func (m *Memory) GetCA(_ context.Context, id string) (model.CA, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ca, ok := m.cas[id]
	if !ok {
		return model.CA{}, ErrNotFound
	}
	return sanitizeCA(ca), nil
}

func (m *Memory) GetCAWithKey(_ context.Context, id string) (model.CA, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ca, ok := m.cas[id]
	if !ok {
		return model.CA{}, ErrNotFound
	}
	return ca, nil
}

func (m *Memory) ListCAs(_ context.Context) ([]model.CA, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	out := make([]model.CA, 0, len(m.cas))
	for _, ca := range m.cas {
		out = append(out, sanitizeCA(ca))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func (m *Memory) UpdateCA(_ context.Context, id string, patch model.CAUpdate) (model.CA, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	ca, ok := m.cas[id]
	if !ok {
		return model.CA{}, ErrNotFound
	}

	if patch.CommonName != nil {
		ca.CommonName = *patch.CommonName
	}
	if patch.Organization != nil {
		ca.Organization = *patch.Organization
	}
	if patch.Country != nil {
		ca.Country = *patch.Country
	}
	if patch.Province != nil {
		ca.Province = *patch.Province
	}
	if patch.Locality != nil {
		ca.Locality = *patch.Locality
	}

	m.cas[id] = ca
	return sanitizeCA(ca), nil
}

func (m *Memory) DeleteCA(_ context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.cas[id]; !ok {
		return ErrNotFound
	}
	delete(m.cas, id)
	delete(m.certs, id)
	return nil
}

func (m *Memory) CreateCert(_ context.Context, caID string, cert model.LeafCert) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.cas[caID]; !ok {
		return ErrNotFound
	}
	if m.certs[caID] == nil {
		m.certs[caID] = make(map[string]model.LeafCert)
	}
	if _, ok := m.certs[caID][cert.ID]; ok {
		return ErrAlreadyExists
	}
	m.certs[caID][cert.ID] = cert
	return nil
}

func (m *Memory) GetCert(_ context.Context, caID, id string) (model.LeafCert, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if _, ok := m.cas[caID]; !ok {
		return model.LeafCert{}, ErrNotFound
	}
	cert, ok := m.certs[caID][id]
	if !ok {
		return model.LeafCert{}, ErrNotFound
	}
	return cert, nil
}

func (m *Memory) ListCerts(_ context.Context, caID string) ([]model.LeafCert, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if _, ok := m.cas[caID]; !ok {
		return nil, ErrNotFound
	}

	out := make([]model.LeafCert, 0, len(m.certs[caID]))
	for _, cert := range m.certs[caID] {
		out = append(out, sanitizeCert(cert))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func (m *Memory) UpdateCert(_ context.Context, caID, id string, cert model.LeafCert) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.cas[caID]; !ok {
		return ErrNotFound
	}
	if _, ok := m.certs[caID][id]; !ok {
		return ErrNotFound
	}
	m.certs[caID][id] = cert
	return nil
}

func (m *Memory) DeleteCert(_ context.Context, caID, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.cas[caID]; !ok {
		return ErrNotFound
	}
	if _, ok := m.certs[caID][id]; !ok {
		return ErrNotFound
	}
	delete(m.certs[caID], id)
	return nil
}

func sanitizeCA(ca model.CA) model.CA {
	ca.KeyPEM = ""
	return ca
}

func sanitizeCert(cert model.LeafCert) model.LeafCert {
	cert.KeyPEM = ""
	return cert
}
