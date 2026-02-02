package main

import (
	"bytes"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"

	"golang.org/x/net/http2/hpack"
)

func main() {
	host := "example.com"

	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:443", host), &tls.Config{
		ServerName: host,
		NextProtos: []string{"h2"},
	})
	if err != nil {
		log.Fatalf("error dialing: %v", err)
	}
	defer conn.Close()

	// Make sure that the server supports HTTP/2
	if _, err := conn.Write([]byte("PRI * HTTP/2.0\r\n\r\nSM\r\n\r\n")); err != nil {
		log.Fatalf("error writing HTTP/2 preamble: %v", err)
	}

	if err := writeFrame(conn, frameTypeSettings, flagEmpty, 0, nil); err != nil {
		log.Fatalf("error writing frame: %v", err)
	}

	var hb bytes.Buffer
	enc := hpack.NewEncoder(&hb)
	headerFields := []hpack.HeaderField{
		{Name: ":method", Value: "GET"},
		{Name: ":scheme", Value: "https"},
		{Name: ":authority", Value: host},
		{Name: ":path", Value: "/"},
	}

	for _, hf := range headerFields {
		if err := enc.WriteField(hf); err != nil {
			log.Fatalf("error encoding header field: %v", err)
		}
	}

	if err := writeFrame(conn, frameTypeHeaders, flagEndStream|flagEndHeaders, 1, hb.Bytes()); err != nil {
		log.Fatalf("error writing frame: %v", err)
	}

	for {
		frame, err := parseFrame(conn)
		if err != nil {
			log.Fatalf("error parsing bytes into frame: %v", err)
		}
		log.Printf("%+v\n", frame)
		if frame.frameType == frameTypeData {
			log.Println(string(frame.payload))
		}
	}
}

type Client struct {
	streams      map[uint32]bool
	lastStreamID uint32
}

func (c *Client) initConn(host string, port int) (net.Conn, error) {
	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", host, port), &tls.Config{
		ServerName: host,
		NextProtos: []string{"h2"},
	})
	if err != nil {
		return nil, err
	}

	// Make sure that the server supports HTTP/2
	if _, err := conn.Write([]byte("PRI * HTTP/2.0\r\n\r\nSM\r\n\r\n")); err != nil {
		conn.Close()
		return nil, fmt.Errorf("error writing HTTP/2 preamble: %v", err)
	}

	if err := writeFrame(conn, frameTypeSettings, flagEmpty, connectionControlStreamID, nil); err != nil {
		conn.Close()
		return nil, fmt.Errorf("error writing frame: %v", err)
	}

	return conn, nil
}

func (c *Client) Get(url string) (*Response, error) {

}

type Response struct {
}

const (
	frameTypeData         byte = 0x0
	frameTypeHeaders      byte = 0x1
	frameTypePriority     byte = 0x2
	frameTypeRSTStream    byte = 0x3
	frameTypeSettings     byte = 0x4
	frameTypePushPromise  byte = 0x5
	frameTypePing         byte = 0x6
	frameTypeGoAway       byte = 0x7
	frameTypeWindowUpdate byte = 0x8
	frameTypeContinuation byte = 0x9

	flagEmpty      byte = 0x0
	flagEndStream  byte = 0x1
	flagACK        byte = 0x1
	flagEndHeaders byte = 0x4
	flagPadded     byte = 0x8
	flagPriority   byte = 0x20

	frameMaxLength   = 1<<24 - 1
	frameMaxStreamID = 1<<31 - 1
	frameHeaderLen   = 9

	connectionControlStreamID = 0
)

// writeFrame ...
func writeFrame(w io.Writer, typ byte, flags byte, streamID uint32, payload []byte) error {
	frame, err := newFrame(typ, flags, streamID, payload)
	if err != nil {
		return fmt.Errorf("error creating frame: %w", err)
	}

	_, err = w.Write(frame.Bytes())
	if err != nil {
		return fmt.Errorf("error writing frame: %w", err)
	}
	return nil
}

type frame struct {
	frameType byte
	flags     byte
	streamID  uint32
	payload   []byte
}

// newFrame ...
func newFrame(typ byte, flags byte, streamID uint32, payload []byte) (*frame, error) {
	if len(payload) > frameMaxLength {
		return nil, fmt.Errorf("error payload length exceeds maximum allowed")
	}
	if streamID > frameMaxStreamID {
		return nil, fmt.Errorf("error stream ID exceeds maximum allowed")
	}

	return &frame{
		frameType: typ,
		flags:     flags,
		streamID:  streamID,
		payload:   payload,
	}, nil
}

// parseFrame consumes bytes from the given reader and parse the bytes into
// a frame struct.
func parseFrame(r io.Reader) (*frame, error) {
	frameHeader := make([]byte, frameHeaderLen)
	n, err := r.Read(frameHeader)
	if err != nil {
		return nil, err
	}
	if n != frameHeaderLen {
		return nil, fmt.Errorf("error frame header too short")
	}

	payloadLen := int(frameHeader[2]) | int(frameHeader[1])<<8 | int(frameHeader[0])<<16
	frameType := frameHeader[3]
	flags := frameHeader[4]
	streamID := binary.BigEndian.Uint32(frameHeader[5:])

	payload := make([]byte, payloadLen)
	n, err = r.Read(payload)
	if err != nil {
		return nil, err
	}
	if n != payloadLen {
		return nil, fmt.Errorf("error invalid payload length in frame header")
	}

	return &frame{
		frameType: frameType,
		flags:     flags,
		streamID:  streamID,
		payload:   payload,
	}, nil
}

func (f *frame) Bytes() []byte {
	payloadLen := len(f.payload)

	b := make([]byte, 0, 9+len(f.payload))
	b = append(b,
		byte(payloadLen>>16),
		byte(payloadLen>>8),
		byte(payloadLen),
		byte(f.frameType),
		byte(f.flags),
		byte(f.streamID>>24),
		byte(f.streamID>>16),
		byte(f.streamID>>8),
		byte(f.streamID),
	)
	b = append(b, f.payload...)
	return b
}

func (f *frame) hasFlag(flag byte) bool {
	return f.flags&flag != 0
}
