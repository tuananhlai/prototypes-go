package resp_test

import (
	"reflect"
	"testing"

	"github.com/tuananhlai/prototypes/my-redis/resp"
)

func TestSerializeBulkString(t *testing.T) {
	expected := []byte("$12\r\ngood morning\r\n")
	got := resp.SerializeBulkString([]byte("good morning"))

	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("expected %v, got %v", expected, got)
	}
}

func TestSerializeArray(t *testing.T) {
	expected := []byte("*2\r\n$5\r\nhello\r\n$5\r\nworld\r\n")
	got := resp.SerializeArray([][]byte{
		[]byte("hello"),
		[]byte("world"),
	})

	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("expected %v, got %v", expected, got)
	}
}
