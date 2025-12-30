package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
)

// Send a HTTPS request to example.com.
func main() {
	conn, err := net.Dial("tcp", "example.com:443")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// By default, Go uses the system certificate pools to verify
	// the target's certificate.
	tlsConn := tls.Client(conn, &tls.Config{
		// The target hostname needs to be passed in so that its certificate
		// can be verified.
		ServerName: "example.com",
	})
	defer tlsConn.Close()

	request := "GET / HTTP/1.1\r\nHost: example.com\r\nConnection: close\r\n\r\n"
	_, err = tlsConn.Write([]byte(request))
	if err != nil {
		log.Fatal(err)
	}

	res, err := io.ReadAll(tlsConn)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(res))
}
