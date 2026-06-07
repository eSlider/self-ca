package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/eSlider/self-ca/internal/ca"
	"github.com/eSlider/self-ca/internal/model"
	"github.com/eSlider/self-ca/internal/store"
)

type Handler struct {
	store store.Store
}

func NewHandler(s store.Store) *Handler {
	return &Handler{store: s}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/cas", h.createCA)
	mux.HandleFunc("GET /api/cas", h.listCAs)
	mux.HandleFunc("GET /api/cas/{id}", h.getCA)
	mux.HandleFunc("PUT /api/cas/{id}", h.updateCA)
	mux.HandleFunc("DELETE /api/cas/{id}", h.deleteCA)

	mux.HandleFunc("POST /api/cas/{caId}/certs", h.createCert)
	mux.HandleFunc("GET /api/cas/{caId}/certs", h.listCerts)
	mux.HandleFunc("GET /api/cas/{caId}/certs/{id}", h.getCert)
	mux.HandleFunc("PUT /api/cas/{caId}/certs/{id}", h.updateCert)
	mux.HandleFunc("DELETE /api/cas/{caId}/certs/{id}", h.deleteCert)

	mux.HandleFunc("GET /api/cas/{id}/download/{file}", h.downloadCA)
	mux.HandleFunc("GET /api/cas/{caId}/certs/{id}/download/{file}", h.downloadCert)
}

func (h *Handler) createCA(w http.ResponseWriter, r *http.Request) {
	var req model.CreateCARequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if strings.TrimSpace(req.CommonName) == "" {
		writeError(w, http.StatusBadRequest, "common_name is required")
		return
	}

	generated, err := ca.GenerateCA(ca.CAOptions{
		CommonName:   req.CommonName,
		Organization: req.Organization,
		Country:      req.Country,
		Province:     req.Province,
		Locality:     req.Locality,
		ValidYears:   req.ValidYears,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	id, err := newID()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	record := model.CA{
		ID:           id,
		CommonName:   generated.Cert.Subject.CommonName,
		Organization: firstOrEmpty(generated.Cert.Subject.Organization),
		Country:      firstOrEmpty(generated.Cert.Subject.Country),
		Province:     firstOrEmpty(generated.Cert.Subject.Province),
		Locality:     firstOrEmpty(generated.Cert.Subject.Locality),
		NotBefore:    generated.Cert.NotBefore,
		NotAfter:     generated.Cert.NotAfter,
		IsCA:         true,
		CertPEM:      string(generated.CertPEM),
		KeyPEM:       string(generated.KeyPEM),
	}

	if err := h.store.CreateCA(r.Context(), record); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, record)
}

func (h *Handler) listCAs(w http.ResponseWriter, r *http.Request) {
	cas, err := h.store.ListCAs(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if cas == nil {
		cas = []model.CA{}
	}
	writeJSON(w, http.StatusOK, cas)
}

func (h *Handler) getCA(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	caRecord, err := h.store.GetCA(r.Context(), id)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	caRecord.KeyPEM = ""
	writeJSON(w, http.StatusOK, caRecord)
}

func (h *Handler) updateCA(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var patch model.CAUpdate
	if err := decodeJSON(r, &patch); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	updated, err := h.store.UpdateCA(r.Context(), id, patch)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	updated.KeyPEM = ""
	writeJSON(w, http.StatusOK, updated)
}

func (h *Handler) deleteCA(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.store.DeleteCA(r.Context(), id); err != nil {
		writeStoreError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) createCert(w http.ResponseWriter, r *http.Request) {
	caID := r.PathValue("caId")
	caRecord, err := h.getCAWithKey(r, caID)
	if err != nil {
		writeStoreError(w, err)
		return
	}

	var req model.CreateCertRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if strings.TrimSpace(req.CommonName) == "" {
		writeError(w, http.StatusBadRequest, "common_name is required")
		return
	}

	caCert, caKey, err := ca.ParseCA([]byte(caRecord.CertPEM), []byte(caRecord.KeyPEM))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	generated, err := ca.IssueLeaf(caCert, caKey, ca.LeafOptions{
		CommonName:  req.CommonName,
		DNSNames:    req.DNSNames,
		IPAddresses: req.IPAddresses,
		ValidYears:  req.ValidYears,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	id, err := newID()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	record := model.LeafCert{
		ID:          id,
		CAID:        caID,
		CommonName:  generated.Cert.Subject.CommonName,
		DNSNames:    generated.Cert.DNSNames,
		IPAddresses: ipStrings(generated.Cert.IPAddresses),
		NotBefore:   generated.Cert.NotBefore,
		NotAfter:    generated.Cert.NotAfter,
		CertPEM:     string(generated.CertPEM),
		KeyPEM:      string(generated.KeyPEM),
	}

	if err := h.store.CreateCert(r.Context(), caID, record); err != nil {
		writeStoreError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, record)
}

func (h *Handler) listCerts(w http.ResponseWriter, r *http.Request) {
	caID := r.PathValue("caId")
	certs, err := h.store.ListCerts(r.Context(), caID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	if certs == nil {
		certs = []model.LeafCert{}
	}
	writeJSON(w, http.StatusOK, certs)
}

func (h *Handler) getCert(w http.ResponseWriter, r *http.Request) {
	caID := r.PathValue("caId")
	id := r.PathValue("id")
	cert, err := h.store.GetCert(r.Context(), caID, id)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, cert)
}

func (h *Handler) updateCert(w http.ResponseWriter, r *http.Request) {
	caID := r.PathValue("caId")
	id := r.PathValue("id")

	existing, err := h.store.GetCert(r.Context(), caID, id)
	if err != nil {
		writeStoreError(w, err)
		return
	}

	caRecord, err := h.getCAWithKey(r, caID)
	if err != nil {
		writeStoreError(w, err)
		return
	}

	var patch model.CertUpdate
	if err := decodeJSON(r, &patch); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	commonName := existing.CommonName
	if patch.CommonName != nil {
		commonName = *patch.CommonName
	}
	dnsNames := existing.DNSNames
	if patch.DNSNames != nil {
		dnsNames = patch.DNSNames
	}
	ipAddresses := existing.IPAddresses
	if patch.IPAddresses != nil {
		ipAddresses = patch.IPAddresses
	}
	validYears := 1
	if patch.ValidYears != nil {
		validYears = *patch.ValidYears
	}

	caCert, caKey, err := ca.ParseCA([]byte(caRecord.CertPEM), []byte(caRecord.KeyPEM))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	generated, err := ca.IssueLeaf(caCert, caKey, ca.LeafOptions{
		CommonName:  commonName,
		DNSNames:    dnsNames,
		IPAddresses: ipAddresses,
		ValidYears:  validYears,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	updated := model.LeafCert{
		ID:          id,
		CAID:        caID,
		CommonName:  generated.Cert.Subject.CommonName,
		DNSNames:    generated.Cert.DNSNames,
		IPAddresses: ipStrings(generated.Cert.IPAddresses),
		NotBefore:   generated.Cert.NotBefore,
		NotAfter:    generated.Cert.NotAfter,
		CertPEM:     string(generated.CertPEM),
		KeyPEM:      string(generated.KeyPEM),
	}

	if err := h.store.UpdateCert(r.Context(), caID, id, updated); err != nil {
		writeStoreError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, updated)
}

func (h *Handler) deleteCert(w http.ResponseWriter, r *http.Request) {
	caID := r.PathValue("caId")
	id := r.PathValue("id")
	if err := h.store.DeleteCert(r.Context(), caID, id); err != nil {
		writeStoreError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) getCAWithKey(r *http.Request, id string) (model.CA, error) {
	return h.store.GetCAWithKey(r.Context(), id)
}

func (h *Handler) downloadCA(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	file := r.PathValue("file")

	caRecord, err := h.store.GetCA(r.Context(), id)
	if err != nil {
		writeStoreError(w, err)
		return
	}

	switch file {
	case "ca.pem", "ca.crt":
		writePEM(w, file, caRecord.CertPEM)
	default:
		writeError(w, http.StatusBadRequest, "supported files: ca.pem, ca.crt")
	}
}

func (h *Handler) downloadCert(w http.ResponseWriter, r *http.Request) {
	caID := r.PathValue("caId")
	id := r.PathValue("id")
	file := r.PathValue("file")

	cert, err := h.store.GetCert(r.Context(), caID, id)
	if err != nil {
		writeStoreError(w, err)
		return
	}

	switch file {
	case "cert.pem", "cert.crt":
		writePEM(w, file, cert.CertPEM)
	case "key.pem", "key.key":
		writePEM(w, file, cert.KeyPEM)
	case "chain.pem":
		caRecord, err := h.store.GetCA(r.Context(), caID)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writePEM(w, file, cert.CertPEM+caRecord.CertPEM)
	default:
		writeError(w, http.StatusBadRequest, "supported files: cert.pem, key.pem, chain.pem")
	}
}

func writePEM(w http.ResponseWriter, filename, content string) {
	w.Header().Set("Content-Type", "application/x-pem-file")
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(content))
}

func decodeJSON(r *http.Request, dst any) error {
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		if errors.Is(err, io.EOF) {
			return errors.New("request body is required")
		}
		return err
	}
	var extra json.RawMessage
	if err := dec.Decode(&extra); err != io.EOF {
		return errors.New("request body must contain a single JSON object")
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, model.ErrorResponse{Error: message})
}

func writeStoreError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, store.ErrNotFound):
		writeError(w, http.StatusNotFound, "not found")
	case errors.Is(err, store.ErrAlreadyExists):
		writeError(w, http.StatusConflict, "already exists")
	default:
		writeError(w, http.StatusInternalServerError, err.Error())
	}
}

func newID() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func firstOrEmpty(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func ipStrings(ips []net.IP) []string {
	if len(ips) == 0 {
		return nil
	}
	out := make([]string, 0, len(ips))
	for _, ip := range ips {
		out = append(out, ip.String())
	}
	return out
}
