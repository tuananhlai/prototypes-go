package main

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
)

const (
	addr = ":8080"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Upgrade") != "websocket" {
			http.Error(w, "not a websocket handshake", http.StatusBadRequest)
			return
		}

		clientKey := r.Header.Get("Sec-WebSocket-Key")
		acceptKey := generateAcceptKey(clientKey)

		hijacker, ok := w.(http.Hijacker)
		if !ok {
			return
		}

		conn, bufrw, err := hijacker.Hijack()
		if err != nil {
			log.Println("failed to hijack connection:", err)
			return
		}
		defer conn.Close()

		bufrw.WriteString("HTTP/1.1 101 Switching Protocols\r\n")
		bufrw.WriteString("Upgrade: websocket\r\n")
		bufrw.WriteString("Connection: Upgrade\r\n")
		fmt.Fprintf(bufrw, "Sec-WebSocket-Accept: %s\r\n", acceptKey)
		bufrw.WriteString("\r\n")
		bufrw.Flush()

		wsConn := NewWebsocketConn(bufrw)
		for {
			payload, err := wsConn.ReadMessage()
			if err != nil {
				log.Println("failed to read message: ", err)
				return
			}

			fmt.Printf("Received: %s\n", string(payload))
		}
	})

	log.Println("starting server on", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		panic(err)
	}
}

func generateAcceptKey(clientKey string) string {
	const guid = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
	h := sha1.New()
	h.Write([]byte(clientKey + guid))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

type WebsocketConn struct {
	connRW *bufio.ReadWriter
}

func NewWebsocketConn(connRW *bufio.ReadWriter) *WebsocketConn {
	return &WebsocketConn{
		connRW: connRW,
	}
}

func (wc *WebsocketConn) ReadMessage() (string, error) {
	header := make([]byte, 2)
	// does the cursor(?) move after `Read` is called?
	_, err := wc.connRW.Read(header)
	if err != nil {
		return "", fmt.Errorf("failed to read header: %w", err)
	}

	// what is a mask key?
	maskKey := make([]byte, 4)
	_, err = wc.connRW.Read(maskKey)
	if err != nil {
		return "", fmt.Errorf("failed to read mask key: %w", err)
	}

	// Why use AND bit operator for the payload len?
	payloadLen := int(header[1] & 0x7F)
	payload := make([]byte, payloadLen)
	_, err = wc.connRW.Read(payload)
	if err != nil {
		return "", fmt.Errorf("failed to read payload: %w", err)
	}

	for i := range payloadLen {
		payload[i] ^= maskKey[i%4]
	}

	return string(payload), nil
}
