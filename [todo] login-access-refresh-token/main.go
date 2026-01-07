package main

import (
	"log"
	"net/http"
)

const (
	addr = ":8080"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /login", func(w http.ResponseWriter, r *http.Request) {})
	mux.HandleFunc("POST /refresh", func(w http.ResponseWriter, r *http.Request) {})
	mux.HandleFunc("GET /me", func(w http.ResponseWriter, r *http.Request) {})

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("unable to start server: %v", err)
	}
}
