package bloomfilter_test

import (
	"testing"

	"github.com/tuananhlai/prototypes/bloom-filter/bloomfilter"
)

func TestMightContain(t *testing.T) {
	filter, err := bloomfilter.New(10, 0.1)
	if err != nil {
		t.Fatalf("error creating bloom filter: %v", err)
	}

	filter.Add("one")
	filter.Add("two")

	if !filter.MightContain("one") || !filter.MightContain("two") {
		t.Error("error might contain")
	}

	if filter.MightContain("three") {
		t.Error("error not contain")
	}
}
