package main

import (
	"log"
	"net/http"
)

const (
	addr = ":8080"
)

func main() {
	http.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	log.Println("start server on", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalln("error starting http server")
	}
}

func streamAudio(w http.ResponseWriter, r *http.Request) {

}
