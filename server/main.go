package main

import (
	"crypto/tls"
	"log"
	"net/http"
)

func main() {
	Authmux := http.NewServeMux()
	httpmux := http.NewServeMux()

	// Register routes defined in spotpass.go
	init_auth_routes(Authmux)
	init_auth_routes(httpmux)
	init_http_routes(httpmux)

	go func() {
		log.Println("Starting HTTP server on :80")
		httpServer := &http.Server{
			Addr:    ":80",
			Handler: httpmux,
		}
		if err := httpServer.ListenAndServe(); err != nil {
			log.Fatalf("Port 80 server crashed: %v", err)
		}
	}()

	log.Println("=== Nintendo Video Server Active ===")
	log.Println("Listening on https://10.0.0.49:443 ...")

	cert := "mitmproxy-ca.pem"
	key := "mitmproxy-ca.pem"

	loadedcert, err := tls.LoadX509KeyPair(cert, key)
	if err != nil {
		log.Fatalf("Failed to load certificate pair: %v", err)
	}

	Authserver := &http.Server{
		Addr:    ":443",
		Handler: Authmux,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS10,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256, // Required by go
				tls.TLS_RSA_WITH_AES_128_CBC_SHA,
				tls.TLS_RSA_WITH_AES_256_CBC_SHA,
				tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA,
			},
			Certificates: []tls.Certificate{loadedcert},
		},
	}
	err = Authserver.ListenAndServeTLS(cert, key)
	if err != nil {
		log.Fatalf("Server crashed: %v", err)
	}
}
