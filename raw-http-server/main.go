package main

import (
	"bufio"
	"io"
	"log"
	"net"
)

const (
	addr = ":8080"
)

func main() {
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("server started on", addr)
	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			log.Println("error accepting connection:", err)
			continue
		}

		log.Println("connection accepted")
		go handleConnection(conn)
	}
}

func handleConnection(conn *net.TCPConn) {
	defer func() {
		err := conn.Close()
		if err != nil {
			log.Println("error closing write:", err)
		}
	}()

	r := bufio.NewReader(conn)

	// we can't use io.ReadAll here. io.ReadAll will wait for io.EOF, which will never come
	// unless the http client closes the connection themselves.
	req, err := r.ReadString('\n')
	if err != nil {
		log.Println("error reading request:", err)
		return
	}
	log.Println(req)

	err = conn.CloseRead()
	if err != nil {
		log.Println("error closing read:", err)
		return
	}

	_, err = io.WriteString(conn, "HTTP/1.1 200 OK\r\n\r\n")
	if err != nil {
		log.Println("error writing response:", err)
		return
	}
}
