// Server code written by OniOkami666
package main

import (
	"log"
	"net/http"
)

func init_routes(mux *http.ServeMux) {
	mux.HandleFunc("/CHECK", Check)
	mux.HandleFunc("/49/", serve_episode)

}

func Check(w http.ResponseWriter, r *http.Request) { // NV sends a CHECK command to here
	log.Println("Incoming CHECK request!")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
func serve_episode(w http.ResponseWriter, r *http.Request) {
	log.Println("Incoming Episode request!", r.URL.Path)

	vids := "../" + r.URL.Path
	http.ServeFile(w, r, vids)
}

