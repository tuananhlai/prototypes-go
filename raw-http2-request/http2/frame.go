package http2

import (
	"encoding/binary"
	"fmt"
	"io"
)

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
