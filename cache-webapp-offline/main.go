package main

import (
	"log"
	"net/http"
)

const (
	addr = ":8080"
)

func main() {
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("GET /", fs)

	log.Println("Server started on", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
