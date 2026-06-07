package store_test

import (
	"context"
	"testing"

	"github.com/eSlider/self-ca/internal/ca"
	"github.com/eSlider/self-ca/internal/model"
	"github.com/eSlider/self-ca/internal/store"
)

func sampleCA(t *testing.T, id string) model.CA {
	t.Helper()
	generated, err := ca.GenerateCA(ca.CAOptions{CommonName: "FS Test CA"})
	if err != nil {
		t.Fatal(err)
	}
	return model.CA{
		ID:         id,
		CommonName: "FS Test CA",
		IsCA:       true,
		CertPEM:    string(generated.CertPEM),
		KeyPEM:     string(generated.KeyPEM),
		NotBefore:  generated.Cert.NotBefore,
		NotAfter:   generated.Cert.NotAfter,
	}
}

func TestFilesystem_CRUD(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	fs := store.NewFilesystem(dir)

	caRecord := sampleCA(t, "ca-1")
	if err := fs.CreateCA(ctx, caRecord); err != nil {
		t.Fatalf("CreateCA: %v", err)
	}

	got, err := fs.GetCA(ctx, "ca-1")
	if err != nil {
		t.Fatalf("GetCA: %v", err)
	}
	if got.KeyPEM != "" {
		t.Fatal("GetCA must strip key")
	}

	withKey, err := fs.GetCAWithKey(ctx, "ca-1")
	if err != nil {
		t.Fatalf("GetCAWithKey: %v", err)
	}
	if withKey.KeyPEM == "" {
		t.Fatal("GetCAWithKey must include key")
	}

	list, err := fs.ListCAs(ctx)
	if err != nil || len(list) != 1 {
		t.Fatalf("ListCAs = %v, %v", list, err)
	}

	generated, _ := ca.GenerateCA(ca.CAOptions{CommonName: "x"})
	caCert, caKey, _ := ca.ParseCA([]byte(caRecord.CertPEM), []byte(caRecord.KeyPEM))
	leaf, err := ca.IssueLeaf(caCert, caKey, ca.LeafOptions{CommonName: "leaf.test"})
	if err != nil {
		t.Fatal(err)
	}
	_ = generated

	cert := model.LeafCert{
		ID:         "cert-1",
		CAID:       "ca-1",
		CommonName: "leaf.test",
		CertPEM:    string(leaf.CertPEM),
		KeyPEM:     string(leaf.KeyPEM),
	}
	if err := fs.CreateCert(ctx, "ca-1", cert); err != nil {
		t.Fatalf("CreateCert: %v", err)
	}

	gotCert, err := fs.GetCert(ctx, "ca-1", "cert-1")
	if err != nil || gotCert.CertPEM == "" {
		t.Fatalf("GetCert: %v", err)
	}

	if err := fs.DeleteCert(ctx, "ca-1", "cert-1"); err != nil {
		t.Fatalf("DeleteCert: %v", err)
	}
	if err := fs.DeleteCA(ctx, "ca-1"); err != nil {
		t.Fatalf("DeleteCA: %v", err)
	}
	if _, err := fs.GetCA(ctx, "ca-1"); err == nil {
		t.Fatal("expected not found after delete")
	}
}

func TestFilesystem_PersistsAcrossInstances(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()

	fs1 := store.NewFilesystem(dir)
	if err := fs1.CreateCA(ctx, sampleCA(t, "persist-ca")); err != nil {
		t.Fatal(err)
	}

	fs2 := store.NewFilesystem(dir)
	got, err := fs2.GetCAWithKey(ctx, "persist-ca")
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if got.CertPEM == "" || got.KeyPEM == "" {
		t.Fatal("expected persisted CA material")
	}
}
