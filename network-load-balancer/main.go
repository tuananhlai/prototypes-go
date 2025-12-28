package main

import (
	"errors"
	"flag"
	"io"
	"log"
	"math/rand/v2"
	"net"
	"strings"
	"sync"
)

const (
	addr = ":8080"
)

// Open 3 different terminals and run the following commands.
// - `go run ./server --addr :8181`
// - `go run ./server --addr :8180`
// - `go run . --servers 127.0.0.1:8180,127.0.0.1:8181`
func main() {
	var serverListStr string
	flag.StringVar(&serverListStr, "servers", "", "comma-separated list of servers")
	flag.Parse()

	if serverListStr == "" {
		log.Fatal("no servers provided")
	}

	servers := strings.Split(serverListStr, ",")

	manager, err := NewManager(servers)
	if err != nil {
		log.Fatalf("error creating manager: %v", err)
	}

	tcpAddr, _ := net.ResolveTCPAddr("tcp", addr)
	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	log.Println("starting load balancer on", addr)
	for {
		clientConn, err := listener.AcceptTCP()
		if err != nil {
			log.Printf("error accepting connection: %v", err)
			continue
		}

		serverConn, err := manager.GetConn()
		if err != nil {
			log.Printf("error getting server connection: %v", err)
			_ = clientConn.Close()
			continue
		}

		go func() {
			defer clientConn.Close()
			defer serverConn.Close()

			var wg sync.WaitGroup

			// io.Copy will block until EOF is reached on src, so one of them
			// needs to be executed on another thread.
			wg.Go(func() {
				// Forward client request to the selected server.
				_, err = io.Copy(serverConn, clientConn)
				if err != nil {
					log.Printf("error copying data to server: %v", err)
				}
				_ = clientConn.CloseWrite()
			})

			wg.Go(func() {
				// Forward server response back to the client.
				_, err = io.Copy(clientConn, serverConn)
				if err != nil {
					log.Printf("error copying data to client: %v", err)
				}
				_ = serverConn.CloseWrite()
			})

			wg.Wait()
		}()
	}
}

type Manager struct {
	serverURLs []string
}

func NewManager(serverURLs []string) (*Manager, error) {
	if len(serverURLs) == 0 {
		return nil, errors.New("no servers provided")
	}

	return &Manager{
		serverURLs: serverURLs,
	}, nil
}

func (m *Manager) GetConn() (*net.TCPConn, error) {
	url := m.getURL()
	return NewTCPConn(url)
}

func (m *Manager) getURL() string {
	return m.serverURLs[rand.IntN(len(m.serverURLs))]
}

func NewTCPConn(url string) (*net.TCPConn, error) {
	tcpAdr, _ := net.ResolveTCPAddr("tcp", url)
	conn, err := net.DialTCP("tcp", nil, tcpAdr)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
