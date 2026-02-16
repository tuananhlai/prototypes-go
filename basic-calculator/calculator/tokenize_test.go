package calculator

import (
	"reflect"
	"testing"
)

func TestTokenize(t *testing.T) {
	testcases := []struct {
		name     string
		arg      string
		expected []token
		wantErr  bool
	}{
		{
			name: "should tokenize number",
			arg:  "123",
			expected: []token{
				{tokenTypeNumber, "123"},
			},
		},
		{
			name: "should tokenize plus expression",
			arg:  "10+3",
			expected: []token{
				{tokenTypeNumber, "10"},
				{tokenTypePlus, ""},
				{tokenTypeNumber, "3"},
			},
			wantErr: false,
		},
		{
			name: "should tokenize minus expression",
			arg:  "3-10",
			expected: []token{
				{tokenTypeNumber, "3"},
				{tokenTypeMinus, ""},
				{tokenTypeNumber, "10"},
			},
			wantErr: false,
		},
		{
			name: "should tokenize multi-operator expressions",
			arg:  "3-1+2",
			expected: []token{
				{tokenTypeNumber, "3"},
				{tokenTypeMinus, ""},
				{tokenTypeNumber, "1"},
				{tokenTypePlus, ""},
				{tokenTypeNumber, "2"},
			},
		},
		{
			name: "should ignore whitespace",
			arg:  "   1 -    3   ",
			expected: []token{
				{tokenTypeNumber, "1"},
				{tokenTypeMinus, ""},
				{tokenTypeNumber, "3"},
			},
		},
		{
			name: "should parse parenthesis",
			arg:  "(1+2)",
			expected: []token{
				{tokenTypeLParen, ""},
				{tokenTypeNumber, "1"},
				{tokenTypePlus, ""},
				{tokenTypeNumber, "2"},
				{tokenTypeRParen, ""},
			},
		},
		{
			name:    "should fail on empty expression string",
			arg:     "",
			wantErr: true,
		},
		{
			name:    "should fail on unrecognized token",
			arg:     "a/3",
			wantErr: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tokenize(tc.arg)
			if (err != nil) != tc.wantErr {
				t.Errorf("tokenize() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("tokenize() got = %v, want %v", got, tc.expected)
			}
		})
	}
}
