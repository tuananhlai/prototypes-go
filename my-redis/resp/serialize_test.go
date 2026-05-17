package resp_test

import (
	"reflect"
	"testing"

	"github.com/tuananhlai/prototypes/my-redis/resp"
)

func TestSerializeBulkString(t *testing.T) {
	expected := []byte("$12\r\ngood morning\r\n")

	got, err := resp.SerializeBulkString([]byte("good morning"))
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("expected %v, got %v", expected, got)
	}
}
