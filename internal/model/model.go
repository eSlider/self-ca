package model

import "time"

type CA struct {
	ID          string    `json:"id"`
	CommonName  string    `json:"common_name"`
	Organization string   `json:"organization,omitempty"`
	Country     string    `json:"country,omitempty"`
	Province    string    `json:"province,omitempty"`
	Locality    string    `json:"locality,omitempty"`
	NotBefore   time.Time `json:"not_before"`
	NotAfter    time.Time `json:"not_after"`
	IsCA        bool      `json:"is_ca"`
	CertPEM     string    `json:"cert_pem,omitempty"`
	KeyPEM      string    `json:"key_pem,omitempty"`
}

type CAUpdate struct {
	CommonName   *string `json:"common_name,omitempty"`
	Organization *string `json:"organization,omitempty"`
	Country      *string `json:"country,omitempty"`
	Province     *string `json:"province,omitempty"`
	Locality     *string `json:"locality,omitempty"`
}

type CreateCARequest struct {
	CommonName   string `json:"common_name"`
	Organization string `json:"organization,omitempty"`
	Country      string `json:"country,omitempty"`
	Province     string `json:"province,omitempty"`
	Locality     string `json:"locality,omitempty"`
	ValidYears   int    `json:"valid_years,omitempty"`
}

type LeafCert struct {
	ID          string    `json:"id"`
	CAID        string    `json:"ca_id"`
	CommonName  string    `json:"common_name"`
	DNSNames    []string  `json:"dns_names,omitempty"`
	IPAddresses []string  `json:"ip_addresses,omitempty"`
	NotBefore   time.Time `json:"not_before"`
	NotAfter    time.Time `json:"not_after"`
	CertPEM     string    `json:"cert_pem,omitempty"`
	KeyPEM      string    `json:"key_pem,omitempty"`
}

type CertUpdate struct {
	CommonName  *string  `json:"common_name,omitempty"`
	DNSNames    []string `json:"dns_names,omitempty"`
	IPAddresses []string `json:"ip_addresses,omitempty"`
	ValidYears  *int     `json:"valid_years,omitempty"`
}

type CreateCertRequest struct {
	CommonName  string   `json:"common_name"`
	DNSNames    []string `json:"dns_names,omitempty"`
	IPAddresses []string `json:"ip_addresses,omitempty"`
	ValidYears  int      `json:"valid_years,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
