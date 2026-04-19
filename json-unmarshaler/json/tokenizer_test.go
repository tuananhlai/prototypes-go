package json

import (
	"reflect"
	"testing"
)

func TestTokenize(t *testing.T) {
	testCases := []struct {
		name      string
		arg       string
		expected  []token
		wantError bool
	}{
		{
			name: "should tokenize {}",
			arg:  "{}",
			expected: []token{
				{
					kind:  TokenOpenParen,
					value: "{",
				},
				{
					kind:  TokenCloseParen,
					value: "}",
				},
			},
			wantError: false,
		},
		{
			name: "should tokenize []",
			arg:  "[]",
			expected: []token{
				{
					kind:  TokenOpenBracket,
					value: "[",
				},
				{
					kind:  TokenCloseBracket,
					value: "]",
				},
			},
			wantError: false,
		},
		{
			name: `should tokenize "foo"`,
			arg:  `"foo"`,
			expected: []token{
				{
					kind:  TokenString,
					value: "foo",
				},
			},
			wantError: false,
		},
		{
			name: `should tokenize 'foo\nbar'`,
			arg:  `"foo\nbar"`,
			expected: []token{
				{
					kind:  TokenString,
					value: "foo\nbar",
				},
			},
		},
		{
			name: `should tokenize 'foo\u2318bar'`,
			arg:  `"foo\u2318bar"`,
			expected: []token{
				{
					kind:  TokenString,
					value: "foo\u2318bar",
				},
			},
		},
		{
			name: "should tokenize 'true'",
			arg:  "true",
			expected: []token{
				{
					kind:  TokenBool,
					value: "true",
				},
			},
			wantError: false,
		},
		{
			name: "should tokenize 'false'",
			arg:  "false",
			expected: []token{
				{
					kind:  TokenBool,
					value: "false",
				},
			},
			wantError: false,
		},
		{
			name: "should tokenize '123'",
			arg:  "123",
			expected: []token{
				{
					kind:  TokenInteger,
					value: "123",
				},
			},
			wantError: false,
		},
		{
			name: `should tokenize '{"key1":"foo","key2":234,"key3":true,"key4":false,"key5":null}'`,
			arg:  `{"key1":"foo","key2":234,"key3":true,"key4":false,"key5":null}`,
			expected: []token{
				{kind: TokenOpenParen, value: "{"},
				{kind: TokenString, value: "key1"},
				{kind: TokenColon, value: ":"},
				{kind: TokenString, value: "foo"},
				{kind: TokenComma, value: ","},
				{kind: TokenString, value: "key2"},
				{kind: TokenColon, value: ":"},
				{kind: TokenInteger, value: "234"},
				{kind: TokenComma, value: ","},
				{kind: TokenString, value: "key3"},
				{kind: TokenColon, value: ":"},
				{kind: TokenBool, value: "true"},
				{kind: TokenComma, value: ","},
				{kind: TokenString, value: "key4"},
				{kind: TokenColon, value: ":"},
				{kind: TokenBool, value: "false"},
				{kind: TokenComma, value: ","},
				{kind: TokenString, value: "key5"},
				{kind: TokenColon, value: ":"},
				{kind: TokenNull, value: "null"},
				{kind: TokenCloseParen, value: "}"},
			},
		},
		{
			name: "should tokenize '3.1415'",
			arg:  "3.1415",
			expected: []token{
				{
					kind:  TokenFloat,
					value: "3.1415",
				},
			},
		},
		{
			name: "should tokenize '0'",
			arg:  "0",
			expected: []token{
				{
					kind:  TokenInteger,
					value: "0",
				},
			},
		},
		{
			name: "should tokenize '0.1234'",
			arg:  "0.1234",
			expected: []token{
				{
					kind:  TokenFloat,
					value: "0.1234",
				},
			},
		},
		{
			name: "should tokenize '1e+01'",
			arg:  "1e+01",
			expected: []token{
				{kind: TokenFloat, value: "1e+01"},
			},
		},
		{
			name: "should tokenize '0.1234e-0123'",
			arg:  "0.1234e-0123",
			expected: []token{
				{kind: TokenFloat, value: "0.1234e-0123"},
			},
		},
		{
			name: "should tokenize '-0.1234e-0123'",
			arg:  "-0.1234e-0123",
			expected: []token{
				{kind: TokenFloat, value: "-0.1234e-0123"},
			},
		},
		{
			name:      "should fail on unrecognized token",
			arg:       "/",
			expected:  nil,
			wantError: true,
		},
		{
			name:      "should fail on empty string",
			arg:       "",
			expected:  nil,
			wantError: true,
		},
		{
			name:      `should fail on '"foo'`,
			arg:       `"foo`,
			expected:  nil,
			wantError: true,
		},
		{
			name:      `should fail on '{}"`,
			arg:       `{}"`,
			expected:  nil,
			wantError: true,
		},
		{
			name:      "should fail on 'tru'",
			arg:       "tru",
			expected:  nil,
			wantError: true,
		},
		{
			name:      "should fail on 'fals'",
			arg:       "fals",
			expected:  nil,
			wantError: true,
		},
		{
			name:      "should fail on '0.'",
			arg:       "0.",
			expected:  nil,
			wantError: true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			tokenizer := newTokenizer(tC.arg)
			tokens, err := tokenizer.tokenize()
			if err != nil && !tC.wantError {
				t.Fatalf("got unexpected error for input %v: %v", tC.arg, err)
			}

			if !reflect.DeepEqual(tokens, tC.expected) {
				t.Errorf("failed on input %v. got: %v, want: %v", tC.arg, tokens, tC.expected)
			}
		})
	}
}
