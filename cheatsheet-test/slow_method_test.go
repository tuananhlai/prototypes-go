package cheatsheettest_test

import (
	"testing"

	cheatsheettest "github.com/tuananhlai/prototypes/cheatsheet-test"
)

func TestSlowMethod(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping slow test in short mode")
	}

	got := cheatsheettest.SlowMethod()
	want := "slow method"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
