package main

import (
	"errors"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"
)

const (
	addr = ":8080"
)

// Define the backend servers to load balance across
var backends = []string{
	"http://localhost:8081",
	"http://localhost:8082",
}

func main() {
	strategy, err := NewRoundRobinStrategy(backends)
	if err != nil {
		log.Fatalf("error creating round robin strategy: %v", err)
	}
	proxy := NewProxy(strategy)

	http.Handle("/", proxy)
	log.Println("Load balancer started on", addr)
	log.Printf("Proxying to backends: %v\n", backends)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("error starting http server: %v", err)
	}
}

type Proxy struct {
	strategy Strategy
}

func NewProxy(strategy Strategy) *Proxy {
	return &Proxy{strategy: strategy}
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	targetURL, err := url.Parse(p.strategy.NextBackend())
	if err != nil {
		http.Error(w, "Bad backend", http.StatusInternalServerError)
		return
	}
	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	proxy.ErrorHandler = func(w http.ResponseWriter, req *http.Request, e error) {
		http.Error(w, "Backend unreachable", http.StatusBadGateway)
	}
	proxy.ServeHTTP(w, r)
}

type Strategy interface {
	NextBackend() string
}

type RoundRobinStrategy struct {
	current  atomic.Uint32
	backends []string
}

func NewRoundRobinStrategy(backends []string) (*RoundRobinStrategy, error) {
	if len(backends) == 0 {
		return nil, errors.New("no backends provided")
	}
	return &RoundRobinStrategy{
		backends: backends,
	}, nil
}

func (r *RoundRobinStrategy) NextBackend() string {
	idx := r.current.Add(1) - 1
	return r.backends[idx%uint32(len(r.backends))]
}
