package main

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	handler := http.StripPrefix("/assets/", http.FileServer(http.Dir("assets")))
	mux.Handle("/assets/", handler)

	log.Println("Starting http server at :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("error starting http server: %v", err)
	}
}
