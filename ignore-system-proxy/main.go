package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptrace"

	"github.com/mattn/go-ieproxy"
)

// Set up a proxy and run the program using:
// `HTTPS_PROXY=[your_proxy_url] go run .
func main() {
	fmt.Println("== Respect proxy environment variables ==")
	demoRespectProxyEnv()
	fmt.Println("== Ignore proxy configuration ==")
	demoIgnoreProxy()
	fmt.Println("== Respect system proxy configuration ==")
	demoRespectSystemProxy()

	fmt.Println("If `Remote Addr` is your proxy URL, it means your proxy was used for that particular request.")
}

func demoRespectProxyEnv() {
	client := &http.Client{
		Transport: &http.Transport{
			// Read proxy address from environment variables like HTTP_PROXY and HTTPS_PROXY
			Proxy: http.ProxyFromEnvironment,
		},
	}

	req := createGetRequest()
	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
}

func demoIgnoreProxy() {
	client := &http.Client{
		Transport: &http.Transport{
			// Explicitly disable proxy.
			Proxy: nil,
		},
	}

	req := createGetRequest()
	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
}

func demoRespectSystemProxy() {
	client := &http.Client{
		Transport: &http.Transport{
			// Fetch proxy configuration from the host OS (works on MacOS and Windows).
			Proxy: ieproxy.GetProxyFunc(),
		},
	}

	req := createGetRequest()
	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
}

func createGetRequest() *http.Request {
	trace := &httptrace.ClientTrace{
		GotConn: func(connInfo httptrace.GotConnInfo) {
			fmt.Printf("--- Connection Trace ---\n")
			fmt.Printf("Remote Addr: %s\n", connInfo.Conn.RemoteAddr())
			fmt.Printf("Was Reused:  %v\n", connInfo.Reused)
			fmt.Printf("Was Idled:   %v\n", connInfo.WasIdle)
			fmt.Printf("------------------------\n\n")
		},
	}

	req, _ := http.NewRequest("GET", "https://example.com", nil)
	return req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
}
