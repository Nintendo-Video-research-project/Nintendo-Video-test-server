package main

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	init_routes(mux)

	log.Println("Running on port 80")
	err := http.ListenAndServe(":80", mux)
	if err != nil {
		log.Fatal(err)
	}
}

