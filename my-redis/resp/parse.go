package resp

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

// ParseArray consumes RESP-encoded bytes from the given reader and construct an array from it.
func ParseArray(r *bufio.Reader) ([][]byte, error) {
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
