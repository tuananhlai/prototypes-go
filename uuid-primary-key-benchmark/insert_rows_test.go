package uuidprimarykeybenchmark_test

import (
	"testing"

	uuidprimarykeybenchmark "github.com/tuananhlai/prototypes/uuid-primary-key-benchmark"
)

func TestInsertOneMilRows(t *testing.T) {
	err := uuidprimarykeybenchmark.InsertOneMilRows()
	if err != nil {
		t.Fatal(err)
	}
}
