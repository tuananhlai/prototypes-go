package main

import (
	"io"
	"log"
	"math/rand"
	"net"
)

const (
	addr = ":8080"
)

func main() {
	servers := []string{"127.0.0.1:8180", "127.0.0.1:8181"}

	conns := make([]net.Conn, 0, len(servers))
	for _, server := range servers {
		conn, err := net.Dial("tcp", server)
		if err != nil {
			log.Fatalf("error connecting to server %s: %v", server, err)
		}

		conns = append(conns, conn)
	}
	defer func() {
		for _, conn := range conns {
			conn.Close()
		}
	}()

	manager := &Manager{
		conns: conns,
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	log.Println("starting load balancer on", addr)
	for {
		clientConn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}

		serverConn := manager.GetConn()
		go func() {
			defer clientConn.Close()

			_, err = io.Copy(serverConn, clientConn)
			if err != nil {
				log.Printf("error copying data to server: %v", err)
			}
			_, err = io.Copy(clientConn, serverConn)
			if err != nil {
				log.Printf("error copying data to client: %v", err)
			}
		}()
	}
}

type Manager struct {
	conns []net.Conn
}

func (m *Manager) GetConn() net.Conn {
	return m.conns[rand.Intn(len(m.conns))]
}
