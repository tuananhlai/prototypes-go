package resp

import (
	"fmt"
	"strconv"
)

var NullBulkString = []byte{respTypeBulkString, '-', '1', '\r', '\n'}

// SerializeBulkString encodes the given bytes into a RESP bulk string. If it is null,
// a serialized null bulk string will be returned.
func SerializeBulkString(b []byte) []byte {
	if b == nil {
		return NullBulkString
	}

	strLen := []byte(strconv.Itoa(len(b)))

	retval := []byte{respTypeBulkString}
	retval = append(retval, strLen...)
	retval = append(retval, '\r', '\n')
	retval = append(retval, b...)
	retval = append(retval, '\r', '\n')

	return retval
}

func SerializeSimpleString(s string) []byte {
	return fmt.Appendf(nil, "+%s\r\n", s)
}

func SerializeSimpleError(s string) []byte {
	return fmt.Appendf(nil, "-%s\r\n", s)
}

func SerializeInteger(v int) []byte {
	return fmt.Appendf(nil, ":%d\r\n", v)
}
