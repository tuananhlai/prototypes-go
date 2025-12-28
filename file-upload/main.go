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

	mux.HandleFunc("POST /uploads:prepare", func(w http.ResponseWriter, r *http.Request) {})
	mux.HandleFunc("POST /uploads/{uploadID}:complete", func(w http.ResponseWriter, r *http.Request) {})

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
