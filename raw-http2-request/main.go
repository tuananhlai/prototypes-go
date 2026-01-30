package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"log"

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

	if err := writeFrame(conn, FrameTypeSettings, FlagEmpty, 0, nil); err != nil {
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

	if err := writeFrame(conn, FrameTypeHeaders, FlagEndStream|FlagEndHeaders, 1, hb.Bytes()); err != nil {
		log.Fatalf("error writing frame: %v", err)
	}

	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		log.Fatalf("error reading response: %v", err)
	}
	log.Printf("%v", buf[:n])

	// TODO: parse HTTP2 response
}

// writeFrame ...
func writeFrame(w io.Writer, typ FrameType, flags Flag, streamID uint32, payload []byte) error {
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

type FrameType byte
type Flag byte

const (
	FrameTypeData         FrameType = 0x0
	FrameTypeHeaders      FrameType = 0x1
	FrameTypePriority     FrameType = 0x2
	FrameTypeRSTStream    FrameType = 0x3
	FrameTypeSettings     FrameType = 0x4
	FrameTypePushPromise  FrameType = 0x5
	FrameTypePing         FrameType = 0x6
	FrameTypeGoAway       FrameType = 0x7
	FrameTypeWindowUpdate FrameType = 0x8
	FrameTypeContinuation FrameType = 0x9

	FlagEmpty      Flag = 0x0
	FlagEndStream  Flag = 0x1
	FlagACK        Flag = 0x1
	FlagEndHeaders Flag = 0x4
	FlagPadded     Flag = 0x8
	FlagPriority   Flag = 0x20

	FrameMaxLength   = 1<<24 - 1
	FrameMaxStreamID = 1<<31 - 1
)

type frame struct {
	length    uint32
	frameType FrameType
	flags     Flag
	streamID  uint32
	payload   []byte
}

// newFrame ...
func newFrame(typ FrameType, flags Flag, streamID uint32, payload []byte) (*frame, error) {
	if len(payload) > FrameMaxLength {
		return nil, fmt.Errorf("error payload length exceeds maximum allowed")
	}
	if streamID > FrameMaxStreamID {
		return nil, fmt.Errorf("error stream ID exceeds maximum allowed")
	}

	return &frame{
		length:    uint32(len(payload)),
		frameType: typ,
		flags:     flags,
		streamID:  streamID,
		payload:   payload,
	}, nil
}

func (f *frame) Bytes() []byte {
	b := make([]byte, 0, 9+len(f.payload))
	b = append(b,
		byte(f.length>>16),
		byte(f.length>>8),
		byte(f.length),
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
