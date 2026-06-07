package store

import (
	"context"
	"errors"

	"github.com/eSlider/self-ca/internal/model"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
)

type Store interface {
	CreateCA(ctx context.Context, ca model.CA) error
	GetCA(ctx context.Context, id string) (model.CA, error)
	GetCAWithKey(ctx context.Context, id string) (model.CA, error)
	ListCAs(ctx context.Context) ([]model.CA, error)
	UpdateCA(ctx context.Context, id string, patch model.CAUpdate) (model.CA, error)
	DeleteCA(ctx context.Context, id string) error

	CreateCert(ctx context.Context, caID string, cert model.LeafCert) error
	GetCert(ctx context.Context, caID, id string) (model.LeafCert, error)
	ListCerts(ctx context.Context, caID string) ([]model.LeafCert, error)
	UpdateCert(ctx context.Context, caID, id string, cert model.LeafCert) error
	DeleteCert(ctx context.Context, caID, id string) error
}
