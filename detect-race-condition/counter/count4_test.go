package counter_test

import (
	"testing"

	"github.com/tuananhlai/prototypes/detect-race-condition/counter"
)

func TestCountToFour(t *testing.T) {
	if counter.CountToFour() != 4 {
		t.Fail()
	}
}
