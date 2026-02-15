package calculator_test

import (
	"fmt"
	"testing"

	"github.com/tuananhlai/prototypes/basic-calculator/calculator"
)

func TestCalculate(t *testing.T) {
	testcases := []struct {
		arg      string
		expected int
		wantErr  bool
	}{
		{
			arg:      "1+2",
			expected: 3,
		},
		{
			arg:      "50-38",
			expected: 12,
		},
		{
			arg:      "3-2+1",
			expected: 2,
		},
		{
			arg:     "1+",
			wantErr: true,
		},
		{
			arg:     "+2",
			wantErr: true,
		},
	}

	for _, tc := range testcases {
		t.Run(fmt.Sprintf("should calculate '%s'", tc.arg), func(t *testing.T) {
			got, err := calculator.Calculate(tc.arg)
			if (err != nil) != tc.wantErr {
				t.Errorf("Calculate() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if got != tc.expected {
				t.Errorf("Calculate() got = %v, want %v", got, tc.expected)
			}
		})
	}
}
