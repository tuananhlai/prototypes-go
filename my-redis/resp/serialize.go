package resp

import (
	"fmt"
)

var NullBulkString = []byte{respTypeBulkString, '-', '1', '\r', '\n'}

// SerializeBulkString encodes the given bytes into a RESP bulk string. If it is null,
// a serialized null bulk string will be returned.
//
// https://redis.io/docs/latest/develop/reference/protocol-spec/#bulk-strings
func SerializeBulkString(b []byte) []byte {
	if b == nil {
		return NullBulkString
	}

	return fmt.Appendf(nil, "%c%d\r\n%s\r\n", respTypeBulkString, len(b), b)
}

// SerializeSimpleString creates a RESP-encoded byte array for the given simple string.
//
// https://redis.io/docs/latest/develop/reference/protocol-spec/#simple-strings
func SerializeSimpleString(s string) []byte {
	return fmt.Appendf(nil, "+%s\r\n", s)
}

// SerializeSimpleError creates a RESP-encoded byte array for the given simple error.
//
// https://redis.io/docs/latest/develop/reference/protocol-spec/#simple-errors
func SerializeSimpleError(s string) []byte {
	return fmt.Appendf(nil, "-%s\r\n", s)
}

// SerializeInteger creates a RESP-encoded byte array for the given integer.
//
// https://redis.io/docs/latest/develop/reference/protocol-spec/#integers
func SerializeInteger(v int) []byte {
	return fmt.Appendf(nil, ":%d\r\n", v)
}

// SerializeArray creates a RESP-encoded byte array from the given string array.
//
// https://redis.io/docs/latest/develop/reference/protocol-spec/#arrays
func SerializeArray(v [][]byte) []byte {
	retval := fmt.Appendf(nil, "%c%d\r\n", respTypeArray, len(v))
	for _, elem := range v {
		retval = append(retval, SerializeBulkString(elem)...)
	}
	return retval
}
