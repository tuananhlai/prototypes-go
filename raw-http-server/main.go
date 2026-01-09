package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
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
	// req, err := r.ReadString('\n')
	// if err != nil {
	// 	log.Println("error reading request:", err)
	// 	return
	// }
	req, err := parse(r)
	if err != nil {
		log.Println("error closing read:", err)
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

type Request struct {
	Method  string
	Path    string
	Headers map[string]string
	Body    []byte
}

func parse(reader *bufio.Reader) (*Request, error) {
	var r Request

	methodLine, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	log.Println(methodLine)

	r.Method, r.Path, err = parseMethodLine(methodLine)
	if err != nil {
		return nil, err
	}

	headers := make(map[string]string, 0)
	for {
		headerLine, err := reader.ReadString('\n')
		if err != nil {
			log.Println("error reading header:", err)
			return nil, err
		}

		if headerLine == "\r\n" {
			break
		}

		key, value, err := parseHeader(headerLine)
		if err != nil {
			log.Println("error parsing header:", err)
			return nil, err
		}
		headers[key] = value
	}
	r.Headers = headers

	return &r, nil
}

func parseMethodLine(methodLine string) (method string, path string, err error) {
	parts := strings.Split(strings.TrimSpace(methodLine), " ")
	if len(parts) != 3 {
		return "", "", fmt.Errorf("invalid method line: %s", methodLine)
	}

	method = parts[0]
	path = parts[1]
	httpVersion := parts[2]

	if httpVersion != "HTTP/1.1" {
		return "", "", fmt.Errorf("unsupported HTTP version: %s", httpVersion)
	}

	return method, path, nil
}

func parseHeader(headerLine string) (key, value string, err error) {
	parts := strings.SplitN(headerLine, ":", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid header line: %s", headerLine)
	}

	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]), nil
}
