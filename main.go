package main

import (
	"crypto/x509"
	"log"
	"net/http"
	"os"
)

func main() {
	certPool := x509.NewCertPool()
	caCert, _ := os.ReadFile("my_ca.crt")
	certPool.AppendCertsFromPEM(caCert)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, HTTPS"))
	})

	err := http.ListenAndServeTLS(":443", "server.crt", "server.key", nil)
	if err != nil {
		log.Fatal(err)
	}
}
