package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
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

	req, err := parse(r)
	if err != nil {
		log.Println("error closing read:", err)
		return
	}
	log.Printf("request: %+v, body: %s", req, string(req.Body()))

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

type Method string

const (
	MethodGet     Method = "GET"
	MethodPost    Method = "POST"
	MethodPut     Method = "PUT"
	MethodDelete  Method = "DELETE"
	MethodPatch   Method = "PATCH"
	MethodOptions Method = "OPTIONS"
)

var (
	validMethods = map[Method]bool{
		MethodGet:     true,
		MethodPost:    true,
		MethodPut:     true,
		MethodDelete:  true,
		MethodPatch:   true,
		MethodOptions: true,
	}
)

type Request struct {
	method  Method
	path    string
	headers map[string]string
	body    []byte
}

func (r *Request) Header(key string) string {
	return r.headers[strings.ToLower(key)]
}

func (r *Request) Method() Method {
	return r.method
}

func (r *Request) Body() []byte {
	return r.body
}

func (r *Request) Path() string {
	return r.path
}

func parse(reader *bufio.Reader) (*Request, error) {
	var r Request

	requestLine, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	log.Println(requestLine)

	r.method, r.path, err = parseRequestLine(requestLine)
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
	r.headers = headers

	contentLength := r.Header("Content-Length")
	if contentLength != "" {
		length, err := strconv.Atoi(contentLength)
		if err != nil {
			return nil, fmt.Errorf("invalid content length: %s", contentLength)
		}

		r.body = make([]byte, length)
		_, err = io.ReadFull(reader, r.body)
		if err != nil {
			return nil, fmt.Errorf("error reading body: %w", err)
		}
	}

	return &r, nil
}

// parseRequestLine ...
// Example input: GET / HTTP/1.1
func parseRequestLine(requestLine string) (method Method, path string, err error) {
	parts := strings.Split(strings.TrimSpace(requestLine), " ")
	if len(parts) != 3 {
		return "", "", fmt.Errorf("invalid method line: %s", requestLine)
	}

	path = parts[1]
	method, err = parseMethod(parts[0])
	if err != nil {
		return "", "", err
	}

	httpVersion := parts[2]
	if httpVersion != "HTTP/1.1" {
		return "", "", fmt.Errorf("unsupported HTTP version: %s", httpVersion)
	}

	return method, path, nil
}

// parseMethod ...
func parseMethod(method string) (Method, error) {
	if !validMethods[Method(method)] {
		return "", fmt.Errorf("invalid method: %s", method)
	}
	return Method(method), nil
}

// parseHeader ...
// Example input: Content-Length: 100
func parseHeader(headerLine string) (key, value string, err error) {
	parts := strings.SplitN(headerLine, ":", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid header line: %s", headerLine)
	}

	// HTTP header name is case-insensitive, so we'll normalize it to lowercase.
	key = strings.ToLower(parts[0])
	// HTTP header name does not allow preceding or trailing whitespace, so
	if strings.Contains(key, " ") {
		return "", "", fmt.Errorf("invalid header name: %s", key)
	}

	value = strings.TrimSpace(parts[1])
	return key, value, nil
}
