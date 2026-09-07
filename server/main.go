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

	server := &http.Server{
		Addr:    ":443",
		Handler: mux,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS10, // Allows TLS 1.1
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_RSA_WITH_AES_128_CBC_SHA,
				tls.TLS_RSA_WITH_AES_256_CBC_SHA,
				tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
				tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			},
			GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
				// Log the incoming SNI hostname to see what the 3DS is requesting
				log.Printf("TLS Handshake requested for SNI: %s", info.ServerName)

				cert, err := tls.LoadX509KeyPair(cert, key)
				if err != nil {
					return nil, err
				}
				return &cert, nil
			},
		},
	}
	err := server.ListenAndServeTLS(cert, key)
	if err != nil {
		log.Fatalf("Server crashed: %v", err)
	}
}
