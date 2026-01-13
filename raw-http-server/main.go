package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
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
	// Set a read deadline to prevent the connection from being kept open
	// indefinitely when the client stops sending requests.
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))

	defer func() {
		err := conn.Close()
		if err != nil {
			log.Println("error closing connection:", err)
		}
	}()

	r := bufio.NewReader(conn)

	for {
		req, err := parse(r)
		if err != nil {
			// EOF means client closed the connection
			if err == io.EOF {
				log.Println("client closed connection")
			} else {
				log.Println("error parsing request:", err)
			}
			return
		}
		log.Printf("request: %+v, body: %s", req, string(req.Body()))

		// In HTTP/1.1, keep-alive is default unless Connection: close is present
		connection := req.Header("connection")
		keepAlive := connection == "keep-alive"

		// Write response with appropriate Connection header
		err = writeResponse(conn, 200, []byte("OK"), keepAlive)
		if err != nil {
			log.Println("error writing response:", err)
			return
		}

		// If client explicitly requested close, close the connection
		if connection == "close" {
			return
		}

		// For keep-alive, continue to read the next request
		// The connection will be closed if:
		// 1. Client sends Connection: close
		// 2. Client closes the connection (EOF)
		// 3. An error occurs during parsing
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

// parse transforms the raw HTTP request into a Request object.
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

// writeResponse writes the HTTP response to the TCP connection.
// If keepAlive is true, includes Connection: keep-alive header.
func writeResponse(conn *net.TCPConn, status int, body []byte, keepAlive bool) error {
	fmt.Fprintf(conn, "HTTP/1.1 %d %s\r\n", status, http.StatusText(status))
	fmt.Fprintf(conn, "Content-Length: %d\r\n", len(body))
	if keepAlive {
		fmt.Fprintf(conn, "Connection: keep-alive\r\n")
	} else {
		fmt.Fprintf(conn, "Connection: close\r\n")
	}
	fmt.Fprintf(conn, "\r\n")

	_, err := conn.Write(body)
	if err != nil {
		return fmt.Errorf("error writing response: %w", err)
	}
	return nil
}
