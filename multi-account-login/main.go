package main

import (
	"log"
	"net/http"
)

const (
	sessionCookieName = "session_id"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	log.Printf("Starting http server on port %s\n", ":8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("error starting http server: %v", err)
	}
}
