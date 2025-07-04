package main

import (
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestHTTPServer(t *testing.T) {
	// Generate certificates for the test
	// We run this in a temporary directory to avoid cluttering the project root
	tempDir, err := ioutil.TempDir("", "cert-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory to generate certs there
	originalWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalWd)

	generateCerts()

	// Create a handler for our test server
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, HTTPS"))
	})

	// Create a new test server with our handler
	server := httptest.NewUnstartedServer(handler)

	// Load the generated server certificate
	cert, err := tls.LoadX509KeyPair("server.crt", "server.key")
	if err != nil {
		t.Fatalf("Failed to load server key pair: %v", err)
	}

	// Configure the test server to use TLS with our certificate
	server.TLS = &tls.Config{Certificates: []tls.Certificate{cert}}
	server.StartTLS() // Start the server
	defer server.Close()

	// Get the client from the test server. It's pre-configured to trust the server's certificate.
	client := server.Client()

	// Make a request to the server
	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("Failed to make HTTPS request: %v", err)
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, but got %d", http.StatusOK, resp.StatusCode)
	}

	// Check the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	expectedBody := "Hello, HTTPS"
	if string(body) != expectedBody {
		t.Errorf("Expected body '%s', but got '%s'", expectedBody, string(body))
	}
}