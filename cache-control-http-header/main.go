package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	// 1. Heuristic (No Header)
	// Behavior: The browser decides how long to cache based on its own logic.
	// To test: Visit, then navigate away and back. If it doesn't change,
	// the browser "guessed" it was safe to store.
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Default (Heuristic). Current time: %s", time.Now().Format(time.RFC1123))
	})

	// 2. Public caching (Max-Age)
	// To view: Navigate to the URL, then click the address bar and hit Enter.
	// The timestamp will remain frozen.
	http.HandleFunc("/public", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=3600")
		fmt.Fprintf(w, "Cacheable for 1 hour. Current time: %s", time.Now().Format(time.RFC1123))
	})

	// 3. No-Store (Security/Privacy)
	// To view: Every single refresh or navigation will update the timestamp.
	// Check DevTools: The "Size" column will never say "(from cache)".
	http.HandleFunc("/private-data", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		fmt.Fprintf(w, "Sensitive: Never cached. Current time: %s", time.Now().Format(time.RFC1123))
	})

	// 4. No-Cache (Revalidation)
	// To view: The browser sends a request every time.
	// The server returns "304 Not Modified" so the timestamp stays frozen,
	// but the Network tab shows a round-trip to the server happened.
	http.HandleFunc("/dynamic", func(w http.ResponseWriter, r *http.Request) {
		etag := "v1-unique-id"
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("ETag", etag)

		if r.Header.Get("If-None-Match") == etag {
			w.WriteHeader(http.StatusNotModified)
			return
		}
		fmt.Fprintf(w, "Must revalidate. Current time: %s", time.Now().Format(time.RFC1123))
	})

	// 5. Stale-While-Revalidate (Performance + Freshness)
	// Behavior: Browser uses a cached version for 10s. For the next 30s, it
	// shows the "stale" version instantly while fetching a fresh one in the background.
	// To view: Refresh after 15s. You'll see the OLD time first, but the
	// NEXT refresh will show the updated time from that background fetch.
	http.HandleFunc("/swr", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=10, stale-while-revalidate=30")
		fmt.Fprintf(w, "SWR Example. Current time: %s", time.Now().Format(time.RFC1123))
	})

	// 6. Private (User-Specific)
	// Behavior: Only the browser can cache this, not a CDN or Proxy.
	// To view: This looks like /public in a single browser, but in a
	// production environment, a CDN would be forced to ignore it.
	http.HandleFunc("/user-profile", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "private, max-age=60")
		fmt.Fprintf(w, "User-specific data. Current time: %s", time.Now().Format(time.RFC1123))
	})

	// 7. Immutable (Long-term Assets)
	// Behavior: Tells the browser the file will NEVER change.
	// TODO: confirm the effect of `immutable`.
	http.HandleFunc("/permanent", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		fmt.Fprintf(w, "Immutable asset. Current time: %s", time.Now().Format(time.RFC1123))
	})

	// Load some HTML which fetches a script to demonstrate the differences in caching
	// for subresources.
	http.HandleFunc("/index.html", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")

		fmt.Fprint(w, `<html><body>
            <h1>Check the console/network tab</h1>
            <script src="/index.js"></script>
        </body></html>`)
	})

	http.HandleFunc("/index.js", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("Server hit for JS at %s\n", time.Now().Format("15:04:05"))
		w.Header().Set("Cache-Control", "public, max-age=31536000")
		fmt.Fprint(w, `console.log("JS loaded");`)
	})

	fmt.Println("Server starting on :8080...")
	http.ListenAndServe(":8080", nil)
}
