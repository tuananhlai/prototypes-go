package main

import (
	"io"
	"log"
	"net/http"
)

const (
	addr = ":8081"
)

func main() {
	data := map[string]string{}
	key := "example"

	http.HandleFunc("GET /data", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("content-type", "text/plain")
		w.Write([]byte(data[key]))
	})
	http.HandleFunc("POST /data", func(w http.ResponseWriter, r *http.Request) {
		rawBody, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read request body", http.StatusBadRequest)
			return
		}

		data[key] = string(rawBody)
		w.Header().Add("content-type", "text/plain")
		w.Write([]byte(data[key]))
	})

	log.Println("Start server at", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("error starting http server: %v", err)
	}
}
