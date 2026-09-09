package main

import (
	"crypto/tls"
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	// Register routes defined in spotpass.go
	init_routes(mux)

	log.Println("=== Nintendo 3DS/2DS Emulation Server Active ===")
	log.Println("Listening on http://10.0.0.49:443 ...")

	cert := "mitmproxy-ca.pem"
	key := "mitmproxy-ca.pem"

	loadedcert, err := tls.LoadX509KeyPair(cert, key)
	if err != nil {
		log.Fatalf("Failed to load certificate pair: %v", err)
	}

	server := &http.Server{
		Addr:    ":443",
		Handler: mux,
		TLSConfig: &tls.Config{
			MinVersion:   tls.VersionTLS10, // Allows TLS 1.1
			Certificates: []tls.Certificate{loadedcert},
		},
	}
	err = server.ListenAndServeTLS(cert, key)
	if err != nil {
		log.Fatalf("Server crashed: %v", err)
	}
}
