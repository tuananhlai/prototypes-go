package main

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
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
			http.Error(w, "invalid data type for response writer", http.StatusInternalServerError)
			return
		}

		conn, bufrw, err := hijacker.Hijack()
		if err != nil {
			http.Error(w, fmt.Sprint("failed to hijack connection:", err), http.StatusInternalServerError)
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

			if err := wsConn.WriteMessage(payload); err != nil {
				log.Println("failed to write message: ", err)
				return
			}
		}
	})

	log.Println("starting server on", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		panic(err)
	}
}

func generateAcceptKey(clientKey string) string {
	// https://datatracker.ietf.org/doc/html/rfc6455#:~:text=%22258EAFA5%2DE914%2D47DA%2D95CA%2DC5AB0DC85B11%22
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

func (wc *WebsocketConn) ReadMessage() ([]byte, error) {
	// Data framing / Packet structure: https://datatracker.ietf.org/doc/html/rfc6455#section-5
	header := make([]byte, 2)
	// does the cursor(?) move after `Read` is called?
	_, err := wc.connRW.Read(header)
	if err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	// A mask key which is used to decode the client payload.
	// It was created to prevent proxies and other intermediaries from mistaking WebSocket
	// traffic for normal HTTP and applying unsafe behavior.
	maskKey := make([]byte, 4)
	_, err = wc.connRW.Read(maskKey)
	if err != nil {
		return nil, fmt.Errorf("failed to read mask key: %w", err)
	}

	// Why use AND bit operator for the payload len?
	payloadLen := int(header[1] & 0x7F)
	payload := make([]byte, payloadLen)
	_, err = wc.connRW.Read(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to read payload: %w", err)
	}

	// Unmask (decode) the client payload.
	for i := range payloadLen {
		payload[i] ^= maskKey[i%4]
	}

	return payload, nil
}

func (wc *WebsocketConn) WriteMessage(message []byte) error {
	// Data framing / Packet structure: https://datatracker.ietf.org/doc/html/rfc6455#section-5
	messageLen := len(message)
	header := []byte{0x81, 0}

	if messageLen <= 125 {
		header[1] = byte(messageLen)
	} else if messageLen <= 65535 {
		header[1] = 126
		lenBuf := make([]byte, 2)
		binary.BigEndian.PutUint16(lenBuf, uint16(messageLen))
		header = append(header, lenBuf...)
	} else {
		header[1] = 127
		lenBuf := make([]byte, 8)
		binary.BigEndian.PutUint64(lenBuf, uint64(messageLen))
		header = append(header, lenBuf...)
	}

	if _, err := wc.connRW.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}
	if _, err := wc.connRW.Write(message); err != nil {
		return fmt.Errorf("failed to write payload: %w", err)
	}

	err := wc.connRW.Flush()
	if err != nil {
		return fmt.Errorf("failed to flush connection write buffer: %w", err)
	}

	return nil
}
