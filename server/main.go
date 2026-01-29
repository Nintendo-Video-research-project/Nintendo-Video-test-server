package main

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	init_routes(mux)

	log.Println("Running on port 8080")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal(err)
	}
}
