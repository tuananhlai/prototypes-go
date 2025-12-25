package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync/atomic"
)

const (
	addr = ":8081"
)

// A web service that allows you to set and retrieve an integer value.
func main() {
	var val = atomic.Int64{}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /value", func(w http.ResponseWriter, r *http.Request) {
		currentValue := val.Load()
		log.Println("GET /value ->", currentValue)
		json.NewEncoder(w).Encode(GetCounterResponse{
			Value: currentValue,
		})
	})
	mux.HandleFunc("PUT /value", func(w http.ResponseWriter, r *http.Request) {
		var req UpdateCounterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		log.Printf("PUT /value %d", req.Value)

		val.Store(req.Value)
		w.WriteHeader(http.StatusOK)
	})

	log.Println("Starting server on", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalln(err)
	}
}

type UpdateCounterRequest struct {
	Value int64 `json:"value"`
}

type GetCounterResponse struct {
	Value int64 `json:"value"`
}
