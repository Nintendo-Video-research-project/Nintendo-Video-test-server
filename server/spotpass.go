package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

func init_routes(mux *http.ServeMux) {
	mux.HandleFunc("/AC/", handleAutoConnect)
	mux.HandleFunc("/nppl/p01/policylist/3/policylist.xml", grab_policyfile)
	mux.HandleFunc("/npul/p01/recv/", Boss_Recv)
	mux.HandleFunc("/reports", handleReports)
	mux.HandleFunc("/1/", handleAppRequests)
	mux.HandleFunc("/logus-p/LogServer_us_live/Upload", handleLogUpload)
	mux.HandleFunc("/pubus-p/", serve_episode)
	mux.HandleFunc("/", handleUnknown)
}
func handleAppRequests(w http.ResponseWriter, r *http.Request) {

	path := r.URL.Path
	logRequestDetails("APP-REQ", r)

	switch {
	case strings.HasSuffix(path, "/CHECK"):
		Check(w, r)
	case strings.Contains(path, "ESE_MD"):
		serve_episode(w, r)
	default:
		handleUnknown(w, r)
	}
}
func handleReports(w http.ResponseWriter, r *http.Request) {
	logRequestDetails("REPORTS", r)

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	_, _ = io.Copy(io.Discard, r.Body)
	defer r.Body.Close()

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Organization", "Nintendo")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
func servePolicyList(w http.ResponseWriter, r *http.Request) {
	r.Header.Del("Proxy-Authorization")
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	http.ServeFile(w, r, "policylist.xml")
}
func Check(w http.ResponseWriter, r *http.Request) { // NV sends a CHECK command to here
	log.Println("Incoming CHECK request!")

	body := []byte("OK")
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Organization", "Nintendo")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(body)))
	w.Header().Set("Connection", "close")
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}
func grab_policyfile(w http.ResponseWriter, r *http.Request) {
	logRequestDetails("POLICY", r)

	data, err := os.ReadFile("policylist.xml")
	if err != nil {
		log.Printf("[POLICY ERROR] Could not read policylist.xml: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.Header().Set("Connection", "close")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
func serve_episode(w http.ResponseWriter, r *http.Request) {
	logRequestDetails("EPISODE", r)

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Connection", "close")

	http.ServeFile(w, r, "./ESE_MD1")
}
func handleLogUpload(w http.ResponseWriter, r *http.Request) {
	logRequestDetails("LOG-UPLOAD", r)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Organization", "Nintendo")
	w.Header().Set("Connection", "close")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
func Boss_Recv(w http.ResponseWriter, r *http.Request) {
	logRequestDetails("BOSS-RECV", r)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Organization", "Nintendo")
	w.Header().Set("Connection", "close")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
func handleUnknown(w http.ResponseWriter, r *http.Request) {
	logRequestDetails("CATCH-ALL", r)
	// Respond 200 OK to CTR Auto-Connect tests hitting "/"

	if r.URL.Path == "/" {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("X-Organization", "Nintendo")
		w.Header().Set("Connection", "close")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
		return
	}
	// Default 404 for unhandled paths
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("404 Not Found"))
}

// Shared helper function to log header details
func logRequestDetails(tag string, r *http.Request) {
	log.Printf("[%s] %s %s from %s", tag, r.Method, r.URL.Path, r.RemoteAddr)
	log.Printf("[%s] User-Agent: %s", tag, r.UserAgent())

	for name, values := range r.Header {
		for _, value := range values {
			log.Printf("   -> Header: %s = %s", name, value)
		}
	}
}
func handleAutoConnect(w http.ResponseWriter, r *http.Request) {
	logRequestDetails("CTR-AC", r)

	// Standard 3DS response for network checks
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Organization", "Nintendo")
	w.Header().Set("Connection", "close")
	w.WriteHeader(http.StatusOK)

	w.Write([]byte("OK"))
}
