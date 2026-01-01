package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	// 1. Public caching: Allows the response to be cached for 1 hour.
	// To view cached response, navigate to the same url instead of pressing 'refresh'.
	http.HandleFunc("/public", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=3600")
		fmt.Fprintf(w, "This response is cacheable for 1 hour. Current time: %s", time.Now().Format(time.RFC1123))
	})

	// 2. No-store: Prevents any caching. The browser must fetch a fresh copy every time.
	http.HandleFunc("/private-data", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		fmt.Fprintf(w, "Sensitive data: This will not be cached. Current time: %s", time.Now().Format(time.RFC1123))
	})

	// 3. No-cache (Must Revalidate): The browser can store a copy,
	// but must check with the server (ETag/Last-Modified) before using it.
	// Unless the user does a hard refresh, browsers will always check whether data is
	// stale using the etag returned by the server.
	http.HandleFunc("/dynamic", func(w http.ResponseWriter, r *http.Request) {
		etag := "v1-unique-id"
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("ETag", etag)

		if r.Header.Get("If-None-Match") == etag {
			w.WriteHeader(http.StatusNotModified)
			return
		}

		fmt.Fprintf(w, "Cached but must revalidate. Current time: %s", time.Now().Format(time.RFC1123))
	})

	fmt.Println("Server starting on :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
