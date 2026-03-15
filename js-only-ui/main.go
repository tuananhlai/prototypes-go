package main

import (
	"log"
	"net/http"
)

const (
	port = ":8080"
)

func main() {
	http.Handle("/", http.FileServer(http.Dir("./static")))

	log.Printf("HTTP server started on %s\n", port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatalf("error initializing http server: %v", err)
	}
}
