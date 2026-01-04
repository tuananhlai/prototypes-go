package cheatsheettest_test

import (
	"testing"

	cheatsheettest "github.com/tuananhlai/prototypes/cheatsheet-test"
)

func FuzzApplyDiscount(f *testing.F) {
	f.Add(100.0, "SAVE10")
	f.Add(50.0, "SAVE50")

	f.Fuzz(func(t *testing.T, basePrice float64, code string) {
		got, err := cheatsheettest.ApplyDiscount(basePrice, code)
		if err != nil {
			return
		}
		if got < 0 {
			t.Errorf("ApplyDiscount(%f, %s) = %f, want >= 0", basePrice, code, got)
		}
	})
}
