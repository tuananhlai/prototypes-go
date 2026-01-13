package main

import (
	"flag"
	"log"
	"net/http"
)

func main() {
	var addr string
	flag.StringVar(&addr, "addr", ":8180", "address to listen on")
	flag.Parse()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("request received: %+v", r)
		w.Write([]byte("Hello, World!"))
	})

	log.Println("starting server on", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("error starting http server: %v", err)
	}
}
