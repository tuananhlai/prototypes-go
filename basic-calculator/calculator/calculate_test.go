package calculator_test

import (
	"testing"

	"github.com/Knetic/govaluate"
	"github.com/tuananhlai/prototypes/basic-calculator/calculator"
)

func TestCalculate(t *testing.T) {
	testcases := []struct {
		name     string
		arg      string
		expected int
		wantErr  bool
	}{
		{
			name:     "should calculate a single number",
			arg:      "10",
			expected: 10,
		},
		{
			name:     "should calculate a single negative number",
			arg:      "-25",
			expected: -25,
		},
		{
			name:     "should calculate a single plus expression",
			arg:      "1+2",
			expected: 3,
		},
		{
			name:     "should calculate a single minus expression",
			arg:      "50-38",
			expected: 12,
		},
		{
			name:     "should calculate a multi-operator expression",
			arg:      "3-2+1",
			expected: 2,
		},
		{
			name:     "should calculate a single parenthesis expression",
			arg:      "(1+2)",
			expected: 3,
		},
		{
			name:     "should calculate a single negative parenthesis expression",
			arg:      "-(2+3)",
			expected: -5,
		},
		{
			name:     "should calculate an expression with multiple parenthesis and operators",
			arg:      "3-(2+1)-(1-2+1)",
			expected: 0,
		},
		{
			name:     "should ignore whitespace",
			arg:      "1 +   2 - (3 +      4)",
			expected: -4,
		},
		{
			name:     "should support nested parenthesis",
			arg:      "((((9))))",
			expected: 9,
		},
		{
			name:    "should fail on empty string",
			arg:     "",
			wantErr: true,
		},
		{
			name:    "should fail for incomplete binary expression",
			arg:     "1+",
			wantErr: true,
		},
		{
			name:    "should fail for invalid unary operator",
			arg:     "+2",
			wantErr: true,
		},
		{
			name:    "should fail when operator is missing",
			arg:     "1 2",
			wantErr: true,
		},
		{
			name:    "should fail when closing parenthesis is missing",
			arg:     "(1 + 2))",
			wantErr: true,
		},
		{
			name:    "should fail when opening parenthesis is missing",
			arg:     "((1 + 2)",
			wantErr: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
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

func BenchmarkCalculate(b *testing.B) {
	s := "(((25 - (7 + 3)) + (14 - (6 - 2))) - ((9 + (4 - 1)) - 8)) + ((30 - (12 + (5 - 3))) - ((7 - 2) + (6 - (4 + 1)))) - (18 - (9 + (2 - (3 + 1))))"

	for b.Loop() {
		_, _ = calculator.Calculate(s)
	}
}

func BenchmarkGovaluate(b *testing.B) {
	s := "(((25 - (7 + 3)) + (14 - (6 - 2))) - ((9 + (4 - 1)) - 8)) + ((30 - (12 + (5 - 3))) - ((7 - 2) + (6 - (4 + 1)))) - (18 - (9 + (2 - (3 + 1))))"

	for b.Loop() {
		expr, _ := govaluate.NewEvaluableExpression(s)
		_, _ = expr.Evaluate(nil)
	}
}
