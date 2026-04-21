package main

import (
	"log"
	"net/http"
	"os"
)

func init_routes(mux *http.ServeMux) { // Initialize the routes in the server when it goes online
	mux.HandleFunc("/1/49/1/CHECK", Check)
	mux.HandleFunc("p01/policylist/3/policylist.xml", grab_policyfile)
	mux.HandleFunc("/1/49/1/", serve_episode)

}

func Check(w http.ResponseWriter, r *http.Request) { // NV sends a CHECK command to here
	log.Println("Incoming CHECK request!")
	w.WriteHeader(http.StatusNotFound) // we want it to return a 404
}
func grab_policyfile(w http.ResponseWriter, r *http.Request) {
	data, err := os.ReadFile("policylist.xml")
	if err != nil {
		log.Printf("Could not read the policy file!  %v", err)
		http.Error(w, "Internal Server Error", 500)
		return
	}
	w.Header().Set("Content-Type", "application/xml")
	w.Write(data)
}
func serve_episode(w http.ResponseWriter, r *http.Request) {
	log.Println("Incoming Episode request!", r.URL.Path)

	// octet streaming for binary streaming
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	Path := "./ESE_MD1"
	http.ServeFile(w, r, Path)
}
