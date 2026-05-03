package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
)

const (
	respTypeArray      = '*'
	respTypeBulkString = '$'
)

func parse(r *bufio.Reader) ([][]byte, error) {
	respType, err := r.ReadByte()
	if err != nil {
		return nil, err
	}

	switch respType {
	case respTypeArray:
		return readArray(r)
	default:
		return nil, fmt.Errorf("error unsupported resp type: %v", respType)
	}
}

// readArray consumes resp-array serialized bytes from the given reader
// and parse it into a Go slice.
func readArray(r *bufio.Reader) ([][]byte, error) {
	line, err := readUntilCRLF(r)
	if err != nil {
		return nil, err
	}

	numElems, err := strconv.Atoi(string(line))
	if err != nil {
		return nil, err
	}

	retval := make([][]byte, 0, numElems)

	var respType byte
	var elem []byte
	for range numElems {
		respType, err = r.ReadByte()
		if err != nil {
			return nil, err
		}

		switch respType {
		case respTypeBulkString:
			elem, err = readBulkString(r)
			if err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("error unsupported resp type: %v", respType)
		}

		retval = append(retval, elem)
	}

	return retval, nil
}

func readBulkString(r *bufio.Reader) ([]byte, error) {
	line, err := readUntilCRLF(r)
	if err != nil {
		return nil, err
	}

	strLen, err := strconv.Atoi(string(line))
	if err != nil {
		return nil, err
	}

	strData := make([]byte, strLen)
	_, err = io.ReadFull(r, strData)
	if err != nil {
		return nil, err
	}

	crlf := make([]byte, 2)
	_, err = io.ReadFull(r, crlf)
	if err != nil {
		return nil, err
	}
	if crlf[0] != '\r' || crlf[1] != '\n' {
		return nil, fmt.Errorf("error expecting CRLF but got %v", crlf)
	}

	return strData, nil
}

// readUntilCRLF consumes bytes from the given reader until we reach a \r\n pair.
// CRLF characters themselves are not included in the return value.
func readUntilCRLF(r *bufio.Reader) ([]byte, error) {
	var retval []byte

	// keep reading until we reach EOF or find a \r\n pair.
	for {
		line, err := r.ReadBytes('\n')
		if err != nil {
			return nil, err
		}

		retval = append(retval, line...)
		if !bytes.HasSuffix(line, []byte{'\r', '\n'}) {
			continue
		}

		return retval[:len(retval)-2], nil
	}
}

// serializeBulkString encodes the given bytes into a RESP bulk string. If it is null,
// a serialized null bulk string will be returned.
func serializeBulkString(b []byte) ([]byte, error) {
	if b == nil {
		return []byte{respTypeBulkString, '-', '1', '\r', '\n'}, nil
	}

	// TODO: return error if len(b) is larger than allowed.
	strLen := []byte(strconv.Itoa(len(b)))

	retval := []byte{respTypeBulkString}
	retval = append(retval, strLen...)
	retval = append(retval, '\r', '\n')
	retval = append(retval, b...)
	retval = append(retval, '\r', '\n')

	return retval, nil
}

func serializeSimpleString(s string) ([]byte, error) {
	// TODO: add length validation
	return fmt.Appendf(nil, "+%s\r\n", s), nil
}
