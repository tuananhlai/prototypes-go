package main

import (
	"fmt"
	"io"
	"log"
	"net"
)

// Send a HTTP request to example.com.
func main() {
	conn, err := net.Dial("tcp", "example.com:80")
	if err != nil {
		log.Fatalf("error creating tcp connection: %v", err)
	}
	defer conn.Close()

	// These two headers are critical for the HTTP request to be processed correctly.
	request := "GET / HTTP/1.1\r\nHost: example.com\r\nConnection: close\r\n\r\n"

	_, err = conn.Write([]byte(request))
	if err != nil {
		log.Fatalf("error writing data to TCP connection: %v", err)
	}

	rawRes, err := io.ReadAll(conn)
	if err != nil {
		log.Fatalf("error reading response: %v", err)
	}

	fmt.Println(string(rawRes))
}
