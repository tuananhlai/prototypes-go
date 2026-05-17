package resp_test

import (
	"bufio"
	"strings"
	"testing"

	"github.com/tuananhlai/prototypes/my-redis/resp"
)

func TestParse(t *testing.T) {
	// "*2\r\n$4\r\nECHO\r\n$5\r\nhello\r\n"
	input := "*2\r\n$4\r\nECHO\r\n$5\r\nhello\r\n"
	r := bufio.NewReader(strings.NewReader(input))

	elems, err := resp.ParseArray(r)
	if err != nil {
		t.Fatal(err)
	}
	if len(elems) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(elems))
	}
	if string(elems[0]) != "ECHO" {
		t.Errorf("expected ECHO, got %q", elems[2])
	}
	if string(elems[1]) != "hello" {
		t.Errorf("expected hello, got %q", elems[3])
	}
}

func TestParseUnsupportedType(t *testing.T) {
	r := bufio.NewReader(strings.NewReader("+OK\r\n"))
	_, err := resp.ParseArray(r)
	if err == nil {
		t.Fatal("expected error for unsupported type")
	}
}
