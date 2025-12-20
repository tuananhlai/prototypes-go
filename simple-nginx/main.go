package main

import (
	"fmt"
	"log"
	"net/http"
)

func requestLogger(w http.ResponseWriter, r *http.Request) {
	// Log the request details to the console
	log.Printf("Received: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)

	// Also log the X-Forwarded-For header if Nginx is passing it
	if realIP := r.Header.Get("X-Forwarded-For"); realIP != "" {
		log.Printf("Real Client IP: %s", realIP)
	}

	fmt.Fprintf(w, "Request logged successfully.\n")
}

func main() {
	http.HandleFunc("/", requestLogger)

	port := ":8080"
	fmt.Printf("Go server starting on %s\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}
