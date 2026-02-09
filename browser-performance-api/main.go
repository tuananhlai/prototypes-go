package main

import (
	"log"
	"net/http"
	"time"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// artificial delay to increase TTFB
		time.Sleep(100 * time.Millisecond)
		http.ServeFile(w, r, "index.html")
	})
	http.HandleFunc("/icon.svg", func(w http.ResponseWriter, r *http.Request) {
		// increase 'load' dom event delay artificially
		time.Sleep(time.Second)
		http.ServeFile(w, r, "icon.svg")
	})

	addr := ":8080"
	log.Printf("listening on http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
