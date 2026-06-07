package api_test

import (
	"bytes"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/eSlider/self-ca/internal/api"
	"github.com/eSlider/self-ca/internal/model"
	"github.com/eSlider/self-ca/internal/store"
)

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	api.NewHandler(store.NewMemory()).Register(mux)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func postJSON(t *testing.T, client *http.Client, url, body string) *http.Response {
	t.Helper()
	resp, err := client.Post(url, "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("POST %s: %v", url, err)
	}
	return resp
}

func doJSON(t *testing.T, client *http.Client, method, url, body string) *http.Response {
	t.Helper()
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		t.Fatalf("%s %s: build request: %v", method, url, err)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, url, err)
	}
	return resp
}

func assertStatus(t *testing.T, resp *http.Response, want int) {
	t.Helper()
	if resp.StatusCode != want {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		t.Fatalf("status = %d, want %d, body = %s", resp.StatusCode, want, string(body))
	}
}

func decodeJSONBody[T any](t *testing.T, resp *http.Response, dst *T) {
	t.Helper()
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(dst); err != nil {
		t.Fatalf("decode json: %v", err)
	}
}

func createTestCA(t *testing.T, client *http.Client, baseURL string) model.CA {
	t.Helper()
	resp := postJSON(t, client, baseURL+"/api/cas", `{"common_name":"Test CA","organization":"Acme"}`)
	assertStatus(t, resp, http.StatusCreated)
	var ca model.CA
	decodeJSONBody(t, resp, &ca)
	if ca.ID == "" {
		t.Fatal("expected CA id")
	}
	if !ca.IsCA {
		t.Fatal("expected is_ca=true")
	}
	if ca.CertPEM == "" || ca.KeyPEM == "" {
		t.Fatal("expected cert_pem and key_pem on create")
	}
	return ca
}

func TestCA_CRUD(t *testing.T) {
	srv := newTestServer(t)
	client := srv.Client()

	t.Run("create_returns_201", func(t *testing.T) {
		resp := postJSON(t, client, srv.URL+"/api/cas", `{"common_name":"Create CA"}`)
		assertStatus(t, resp, http.StatusCreated)

		var got model.CA
		decodeJSONBody(t, resp, &got)
		if got.CommonName != "Create CA" {
			t.Fatalf("common_name = %q", got.CommonName)
		}
		if !got.IsCA {
			t.Fatal("expected is_ca=true")
		}
		block, _ := pem.Decode([]byte(got.CertPEM))
		if block == nil {
			t.Fatal("invalid cert_pem")
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			t.Fatalf("parse cert: %v", err)
		}
		if !cert.IsCA {
			t.Fatal("parsed cert should be CA")
		}
	})

	t.Run("list_after_create", func(t *testing.T) {
		_ = createTestCA(t, client, srv.URL)
		resp, err := client.Get(srv.URL + "/api/cas")
		if err != nil {
			t.Fatalf("GET list: %v", err)
		}
		assertStatus(t, resp, http.StatusOK)

		var list []model.CA
		decodeJSONBody(t, resp, &list)
		if len(list) == 0 {
			t.Fatal("expected at least one CA")
		}
		for _, item := range list {
			if item.KeyPEM != "" {
				t.Fatal("list must not include key_pem")
			}
		}
	})

	t.Run("get_by_id", func(t *testing.T) {
		ca := createTestCA(t, client, srv.URL)
		resp, err := client.Get(srv.URL + "/api/cas/" + ca.ID)
		if err != nil {
			t.Fatalf("GET: %v", err)
		}
		assertStatus(t, resp, http.StatusOK)

		var got model.CA
		decodeJSONBody(t, resp, &got)
		if got.CertPEM == "" {
			t.Fatal("expected cert_pem")
		}
		if got.KeyPEM != "" {
			t.Fatal("get must not include key_pem")
		}
		if got.CertPEM == "" {
			t.Fatal("expected cert_pem for download")
		}
	})

	t.Run("download_ca_pem", func(t *testing.T) {
		ca := createTestCA(t, client, srv.URL)
		resp, err := client.Get(srv.URL + "/api/cas/" + ca.ID + "/download/ca.pem")
		if err != nil {
			t.Fatal(err)
		}
		assertStatus(t, resp, http.StatusOK)
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if !strings.Contains(string(body), "BEGIN CERTIFICATE") {
			t.Fatal("expected PEM body")
		}
	})

	t.Run("get_missing_404", func(t *testing.T) {
		resp, err := client.Get(srv.URL + "/api/cas/does-not-exist")
		if err != nil {
			t.Fatalf("GET: %v", err)
		}
		assertStatus(t, resp, http.StatusNotFound)
	})

	t.Run("update_metadata", func(t *testing.T) {
		ca := createTestCA(t, client, srv.URL)
		resp := doJSON(t, client, http.MethodPut, srv.URL+"/api/cas/"+ca.ID, `{"common_name":"Updated CA"}`)
		assertStatus(t, resp, http.StatusOK)

		var got model.CA
		decodeJSONBody(t, resp, &got)
		if got.CommonName != "Updated CA" {
			t.Fatalf("common_name = %q", got.CommonName)
		}
		if got.CertPEM != ca.CertPEM {
			t.Fatal("cert PEM should remain unchanged on metadata update")
		}
	})

	t.Run("delete_204", func(t *testing.T) {
		ca := createTestCA(t, client, srv.URL)
		resp := doJSON(t, client, http.MethodDelete, srv.URL+"/api/cas/"+ca.ID, "")
		assertStatus(t, resp, http.StatusNoContent)

		resp, err := client.Get(srv.URL + "/api/cas/" + ca.ID)
		if err != nil {
			t.Fatalf("GET after delete: %v", err)
		}
		assertStatus(t, resp, http.StatusNotFound)
	})

	t.Run("create_invalid_400", func(t *testing.T) {
		resp := postJSON(t, client, srv.URL+"/api/cas", `{"common_name":""}`)
		assertStatus(t, resp, http.StatusBadRequest)
	})
}

func TestCert_CRUD(t *testing.T) {
	srv := newTestServer(t)
	client := srv.Client()
	ca := createTestCA(t, client, srv.URL)

	t.Run("create_under_ca", func(t *testing.T) {
		resp := postJSON(t, client, srv.URL+"/api/cas/"+ca.ID+"/certs",
			`{"common_name":"localhost","dns_names":["localhost"],"ip_addresses":["127.0.0.1"]}`)
		assertStatus(t, resp, http.StatusCreated)

		var cert model.LeafCert
		decodeJSONBody(t, resp, &cert)
		if cert.ID == "" {
			t.Fatal("expected cert id")
		}
		block, _ := pem.Decode([]byte(cert.CertPEM))
		if block == nil {
			t.Fatal("invalid cert_pem")
		}
		parsed, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			t.Fatalf("parse cert: %v", err)
		}
		if parsed.IsCA {
			t.Fatal("leaf cert must not be CA")
		}
	})

	t.Run("create_ca_missing_404", func(t *testing.T) {
		resp := postJSON(t, client, srv.URL+"/api/cas/missing/certs", `{"common_name":"x"}`)
		assertStatus(t, resp, http.StatusNotFound)
	})

	t.Run("list_certs", func(t *testing.T) {
		caRecord := createTestCA(t, client, srv.URL)
		for _, body := range []string{
			`{"common_name":"a.example"}`,
			`{"common_name":"b.example"}`,
		} {
			resp := postJSON(t, client, srv.URL+"/api/cas/"+caRecord.ID+"/certs", body)
			assertStatus(t, resp, http.StatusCreated)
		}

		resp, err := client.Get(srv.URL + "/api/cas/" + caRecord.ID + "/certs")
		if err != nil {
			t.Fatalf("GET list: %v", err)
		}
		assertStatus(t, resp, http.StatusOK)

		var list []model.LeafCert
		decodeJSONBody(t, resp, &list)
		if len(list) != 2 {
			t.Fatalf("len(list) = %d, want 2", len(list))
		}
	})

	t.Run("get_cert_includes_key", func(t *testing.T) {
		caRecord := createTestCA(t, client, srv.URL)
		createResp := postJSON(t, client, srv.URL+"/api/cas/"+caRecord.ID+"/certs", `{"common_name":"secure.local"}`)
		assertStatus(t, createResp, http.StatusCreated)
		var created model.LeafCert
		decodeJSONBody(t, createResp, &created)

		resp, err := client.Get(srv.URL + "/api/cas/" + caRecord.ID + "/certs/" + created.ID)
		if err != nil {
			t.Fatalf("GET cert: %v", err)
		}
		assertStatus(t, resp, http.StatusOK)

		var got model.LeafCert
		decodeJSONBody(t, resp, &got)
		if got.CertPEM == "" || got.KeyPEM == "" {
			t.Fatal("expected cert_pem and key_pem")
		}
	})

	t.Run("update_reissues", func(t *testing.T) {
		caRecord := createTestCA(t, client, srv.URL)
		createResp := postJSON(t, client, srv.URL+"/api/cas/"+caRecord.ID+"/certs",
			`{"common_name":"old.example","dns_names":["old.example"]}`)
		assertStatus(t, createResp, http.StatusCreated)
		var created model.LeafCert
		decodeJSONBody(t, createResp, &created)

		updateResp := doJSON(t, client, http.MethodPut, srv.URL+"/api/cas/"+caRecord.ID+"/certs/"+created.ID,
			`{"dns_names":["new.example"]}`)
		assertStatus(t, updateResp, http.StatusOK)

		var updated model.LeafCert
		decodeJSONBody(t, updateResp, &updated)
		if updated.ID != created.ID {
			t.Fatalf("id changed: %q -> %q", created.ID, updated.ID)
		}
		if updated.CertPEM == created.CertPEM {
			t.Fatal("cert PEM should change after re-issue")
		}
		if len(updated.DNSNames) != 1 || updated.DNSNames[0] != "new.example" {
			t.Fatalf("dns_names = %v", updated.DNSNames)
		}
	})

	t.Run("delete_cert", func(t *testing.T) {
		caRecord := createTestCA(t, client, srv.URL)
		createResp := postJSON(t, client, srv.URL+"/api/cas/"+caRecord.ID+"/certs", `{"common_name":"delete.me"}`)
		assertStatus(t, createResp, http.StatusCreated)
		var created model.LeafCert
		decodeJSONBody(t, createResp, &created)

		delResp := doJSON(t, client, http.MethodDelete, srv.URL+"/api/cas/"+caRecord.ID+"/certs/"+created.ID, "")
		assertStatus(t, delResp, http.StatusNoContent)

		resp, err := client.Get(srv.URL + "/api/cas/" + caRecord.ID + "/certs/" + created.ID)
		if err != nil {
			t.Fatalf("GET after delete: %v", err)
		}
		assertStatus(t, resp, http.StatusNotFound)
	})

	t.Run("delete_ca_cascades", func(t *testing.T) {
		caRecord := createTestCA(t, client, srv.URL)
		var certIDs []string
		for _, cn := range []string{"one.local", "two.local"} {
			resp := postJSON(t, client, srv.URL+"/api/cas/"+caRecord.ID+"/certs",
				`{"common_name":"`+cn+`"}`)
			assertStatus(t, resp, http.StatusCreated)
			var cert model.LeafCert
			decodeJSONBody(t, resp, &cert)
			certIDs = append(certIDs, cert.ID)
		}

		delResp := doJSON(t, client, http.MethodDelete, srv.URL+"/api/cas/"+caRecord.ID, "")
		assertStatus(t, delResp, http.StatusNoContent)

		for _, certID := range certIDs {
			resp, err := client.Get(srv.URL + "/api/cas/" + caRecord.ID + "/certs/" + certID)
			if err != nil {
				t.Fatalf("GET cert after CA delete: %v", err)
			}
			assertStatus(t, resp, http.StatusNotFound)
		}
	})

	t.Run("download_cert_and_chain", func(t *testing.T) {
		caRecord := createTestCA(t, client, srv.URL)
		createResp := postJSON(t, client, srv.URL+"/api/cas/"+caRecord.ID+"/certs", `{"common_name":"dl.test"}`)
		assertStatus(t, createResp, http.StatusCreated)
		var created model.LeafCert
		decodeJSONBody(t, createResp, &created)

		for _, path := range []string{"cert.pem", "key.pem", "chain.pem"} {
			resp, err := client.Get(srv.URL + "/api/cas/" + caRecord.ID + "/certs/" + created.ID + "/download/" + path)
			if err != nil {
				t.Fatalf("GET %s: %v", path, err)
			}
			assertStatus(t, resp, http.StatusOK)
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			if !strings.Contains(string(body), "BEGIN") {
				t.Fatalf("%s: expected PEM", path)
			}
		}
	})

	_ = ca
}

func TestExportCA(t *testing.T) {
	srv := newTestServer(t)
	client := srv.Client()
	caRecord := createTestCA(t, client, srv.URL)

	tests := []struct {
		platform string
		contains string
	}{
		{"mobileconfig", "com.apple.security.root"},
		{"windows-ps1", "Import-Certificate"},
		{"windows-bat", "certutil"},
		{"linux", "update-ca-certificates"},
		{"android", "network-security-config"},
	}

	for _, tc := range tests {
		t.Run(tc.platform, func(t *testing.T) {
			resp, err := client.Get(srv.URL + "/api/cas/" + caRecord.ID + "/export/" + tc.platform)
			if err != nil {
				t.Fatal(err)
			}
			assertStatus(t, resp, http.StatusOK)
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			if !strings.Contains(string(body), tc.contains) {
				t.Fatalf("body missing %q", tc.contains)
			}
		})
	}

	resp, err := client.Get(srv.URL + "/api/cas/" + caRecord.ID + "/export/unknown")
	if err != nil {
		t.Fatal(err)
	}
	assertStatus(t, resp, http.StatusBadRequest)
}

func TestCreateCA_RejectsUnknownFields(t *testing.T) {
	srv := newTestServer(t)
	resp := postJSON(t, srv.Client(), srv.URL+"/api/cas", `{"common_name":"X","extra":true}`)
	assertStatus(t, resp, http.StatusBadRequest)
}

func TestCreateCert_RejectsEmptyBody(t *testing.T) {
	srv := newTestServer(t)
	ca := createTestCA(t, srv.Client(), srv.URL)

	req, err := http.NewRequest(http.MethodPost, srv.URL+"/api/cas/"+ca.ID+"/certs", bytes.NewReader(nil))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	assertStatus(t, resp, http.StatusBadRequest)
}
