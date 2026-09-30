package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const episodeRoot = "C:/Nintendo-Video-test-server-main"

func init_auth_routes(mux *http.ServeMux) {
	mux.HandleFunc("/check", Check)
	mux.HandleFunc("/AC/", handleAutoConnect)
	mux.HandleFunc("/p01/policylist/3/", grab_policyfile)
	mux.HandleFunc("/p01/recv/", Boss_Recv)
	mux.HandleFunc("/p01/nsa/", serve_episode)
	mux.HandleFunc("/reports", handleReports)
	mux.HandleFunc("/1/", handleAppRequests)
	mux.HandleFunc("/", handleUnknown)
}
func init_http_routes(mux *http.ServeMux) {
	mux.HandleFunc("/pubus-p/", serve_episode)
	mux.HandleFunc("/pubeu-p/", serve_episode)
	mux.HandleFunc("/pubes-p/", serve_episode)
}
func handleAppRequests(w http.ResponseWriter, r *http.Request) {

	path := r.URL.Path
	logRequestDetails("APP-REQ", r)

	switch {
	case strings.HasSuffix(path, "/check"):
		Check(w, r)
	case strings.Contains(path, "ESE_MD"):
		serve_episode(w, r)
	default:
		handleUnknown(w, r)
	}
}
func handleReports(w http.ResponseWriter, r *http.Request) {
	logRequestDetails("REPORTS", r)

	if r.Method == http.MethodPost {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("[REPORTS ERROR] Failed to read request body: %v", err)
		} else if len(body) > 0 {
			log.Printf("[REPORTS] Payload received (%d bytes)", len(body))
			_ = os.WriteFile(fmt.Sprintf("report_%s.bin", r.Header.Get("X-Boss-Uniqueid")), body, 0644)
		}
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Organization", "Nintendo")
	// w.Header().Set("Connection", "close")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
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

type statusResponseWriter struct {
	http.ResponseWriter
	status int
}

// Capture the real status code
func (s *statusResponseWriter) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

func serve_episode(w http.ResponseWriter, r *http.Request) {
	logRequestDetails("EPISODE", r)

	sw := &statusResponseWriter{ResponseWriter: w, status: http.StatusOK}

	lowerPath := strings.ToLower(r.URL.Path)
	if strings.Contains(lowerPath, "/check") || strings.HasSuffix(lowerPath, "/check") {
		sw.WriteHeader(http.StatusOK)
		sw.Write([]byte("OK"))
		log.Printf("[EPISODE] Handled CHECK request successfully for: %s", r.URL.Path)
		return
	}

	cleanPath := strings.TrimPrefix(r.URL.Path, "/")
	target := filepath.Join(episodeRoot, cleanPath)

	log.Printf("[EPISODE] Looking for file on disk: %s", target)

	info, err := os.Stat(target)
	if err != nil {
		log.Printf("[EPISODE ERROR] Cannot stat %s: %v", target, err)
		if os.IsNotExist(err) {
			http.Error(sw, "File not found", http.StatusNotFound)
		} else {
			http.Error(sw, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}
	if info.IsDir() {
		http.Error(sw, "File not found", http.StatusNotFound)
		return
	}

	sw.Header().Set("Content-Type", "application/octet-stream")
	sw.Header().Set("X-Organization", "Nintendo")

	http.ServeFile(sw, r, target)
	log.Printf("[EPISODE] Finished serving %s with HTTP Status: %d", cleanPath, sw.status)
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

	if strings.HasSuffix(r.URL.Path, "/sendcfg") {
		body, err := io.ReadAll(r.Body)
		if err == nil && len(body) > 0 {
			log.Printf("[SENDCFG] Received config payload size: %d bytes", len(body))
		}
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Organization", "Nintendo")
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
