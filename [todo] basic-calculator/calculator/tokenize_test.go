package calculator

import (
	"fmt"
	"reflect"
	"testing"
)

func TestTokenize(t *testing.T) {
	testcases := []struct {
		arg      string
		expected []token
		wantErr  bool
	}{
		{
			arg: "1+2",
			expected: []token{
				{tokenTypeNumber, "1"},
				{tokenTypePlus, ""},
				{tokenTypeNumber, "2"},
			},
			wantErr: false,
		},
		{
			arg: "10+39",
			expected: []token{
				{tokenTypeNumber, "10"},
				{tokenTypePlus, ""},
				{tokenTypeNumber, "39"},
			},
			wantErr: false,
		},
		{
			arg: "3-1",
			expected: []token{
				{tokenTypeNumber, "3"},
				{tokenTypeMinus, ""},
				{tokenTypeNumber, "1"},
			},
			wantErr: false,
		},
		{
			arg: "3-1+2",
			expected: []token{
				{tokenTypeNumber, "3"},
				{tokenTypeMinus, ""},
				{tokenTypeNumber, "1"},
				{tokenTypePlus, ""},
				{tokenTypeNumber, "2"},
			},
			wantErr: false,
		},
	}

	for _, tc := range testcases {
		t.Run(fmt.Sprintf("should tokenize '%s'", tc.arg), func(t *testing.T) {
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
