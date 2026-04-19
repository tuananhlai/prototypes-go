package json

import (
	"strconv"
	"strings"
)

type TokenKind int

const (
	TokenOpenParen TokenKind = iota
	TokenCloseParen
	TokenOpenBracket
	TokenCloseBracket
	TokenColon
	TokenString
	TokenBool
	TokenInteger
	TokenFloat
	TokenNull
	TokenComma
)

var invalidCharacterSet = newByteSet('\t', '\b', '\f', '\n', '\r')

type token struct {
	kind  TokenKind
	value string
}

type tokenizer struct {
	input string
	pos   int
}

func newTokenizer(input string) *tokenizer {
	return &tokenizer{
		input: input,
		pos:   0,
	}
}

// tokenize breaks the input JSON string into a list of predefined tokens.
func (t *tokenizer) tokenize() ([]token, error) {
	if len(t.input) == 0 {
		return nil, ErrUnexpectedEOF
	}

	var tokens []token

	for {
		if t.pos > len(t.input)-1 {
			break
		}

		cur := t.input[t.pos]
		switch cur {
		case ' ', '\t', '\n', '\r':
			t.pos++
		case '{':
			t.pos++
			tokens = append(tokens, token{
				kind:  TokenOpenParen,
				value: string(cur),
			})
		case '}':
			t.pos++
			tokens = append(tokens, token{
				kind:  TokenCloseParen,
				value: string(cur),
			})
		case '[':
			t.pos++
			tokens = append(tokens, token{
				kind:  TokenOpenBracket,
				value: string(cur),
			})
		case ']':
			t.pos++
			tokens = append(tokens, token{
				kind:  TokenCloseBracket,
				value: string(cur),
			})
		case ',':
			t.pos++
			tokens = append(tokens, token{
				kind:  TokenComma,
				value: string(cur),
			})
		case ':':
			t.pos++
			tokens = append(tokens, token{
				kind:  TokenColon,
				value: string(cur),
			})
		case '"':
			token, err := t.readString()
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, token)
		case 't', 'f':
			token, err := t.readBoolean()
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, token)
		case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			token, err := t.readNumber()
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, token)
		case 'n':
			token, err := t.readNull()
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, token)
		default:
			return nil, NewUnexpectedTokenError(t.pos, cur)
		}
	}

	return tokens, nil
}

func (t *tokenizer) readString() (token, error) {
	// skip the current " character
	t.pos++
	builder := strings.Builder{}

	var cur byte
	for {
		if t.pos > len(t.input)-1 {
			return token{}, ErrUnexpectedEOF
		}
		cur = t.input[t.pos]

		if invalidCharacterSet.has(cur) {
			return token{}, NewInvalidCharacterError(t.pos, cur)
		}

		if cur == '\\' {
			escapedChar, err := t.readEscapedCharacter()
			if err != nil {
				return token{}, err
			}

			builder.WriteRune(escapedChar)
			continue
		}

		if cur == '"' {
			t.pos++
			return token{
				kind:  TokenString,
				value: builder.String(),
			}, nil
		}

		builder.WriteByte(cur)
		t.pos++
	}
}

func (t *tokenizer) readBoolean() (token, error) {
	switch t.input[t.pos] {
	case 't':
		if t.pos+4 > len(t.input) {
			return token{}, ErrUnexpectedEOF
		}
		if t.input[t.pos:t.pos+4] != "true" {
			return token{}, NewUnexpectedTokenError(t.pos, t.input[t.pos:t.pos+4])
		}

		t.pos += 4
		return token{
			kind:  TokenBool,
			value: "true",
		}, nil
	case 'f':
		if t.pos+5 > len(t.input) {
			return token{}, ErrUnexpectedEOF
		}
		if t.input[t.pos:t.pos+5] != "false" {
			return token{}, NewUnexpectedTokenError(t.pos, t.input[t.pos:t.pos+5])
		}

		t.pos += 5
		return token{
			kind:  TokenBool,
			value: "false",
		}, nil
	default:
		return token{}, ErrUnknownState
	}
}

func (t *tokenizer) readNumber() (token, error) {
	startPos := t.pos
	builder := strings.Builder{}
	tokenKind := TokenInteger

	if t.input[t.pos] == '-' {
		builder.WriteByte('-')
		t.pos++
	}

	if t.pos > len(t.input)-1 {
		return token{}, ErrUnexpectedEOF
	}

	switch t.input[t.pos] {
	case '0':
		builder.WriteByte('0')
		t.pos++

	case '1', '2', '3', '4', '5', '6', '7', '8', '9':
		digits := t.readDigits()
		builder.WriteString(digits)
	default:
		return token{}, NewUnexpectedTokenError(t.pos, t.input[t.pos])
	}

	if t.pos > len(t.input)-1 {
		return token{
			kind:  tokenKind,
			value: builder.String(),
		}, nil
	}

	// handle decimal point
	if t.input[t.pos] == '.' {
		tokenKind = TokenFloat
		builder.WriteByte('.')
		t.pos++
		digits := t.readDigits()
		if len(digits) == 0 {
			return token{}, NewInvalidNumberError(startPos, "no digit found after decimal point")
		}
		builder.WriteString(digits)
	}

	if t.pos > len(t.input)-1 {
		return token{
			kind:  tokenKind,
			value: builder.String(),
		}, nil
	}

	// handle scientific notation
	if t.input[t.pos] == 'e' || t.input[t.pos] == 'E' {
		tokenKind = TokenFloat
		builder.WriteByte(t.input[t.pos])
		t.pos++

		if t.pos > len(t.input)-1 {
			return token{}, ErrUnexpectedEOF
		}

		if t.input[t.pos] == '-' || t.input[t.pos] == '+' {
			builder.WriteByte(t.input[t.pos])
			t.pos++
		}

		digits := t.readDigits()
		if len(digits) == 0 {
			return token{}, NewInvalidNumberError(startPos, "no digit found after exponent")
		}

		builder.WriteString(digits)
	}

	return token{
		kind:  tokenKind,
		value: builder.String(),
	}, nil
}

func (t *tokenizer) readDigits() string {
	builder := strings.Builder{}
	for {
		if t.pos > len(t.input)-1 || t.input[t.pos] < '0' || t.input[t.pos] > '9' {
			return builder.String()
		}
		builder.WriteByte(t.input[t.pos])
		t.pos++
	}
}

func (t *tokenizer) readNull() (token, error) {
	if t.pos+4 > len(t.input) {
		return token{}, ErrUnexpectedEOF
	}

	if t.input[t.pos:t.pos+4] != "null" {
		return token{}, NewUnexpectedTokenError(t.pos, t.input[t.pos:t.pos+4])
	}

	t.pos += 4
	return token{
		kind:  TokenNull,
		value: "null",
	}, nil
}

func (t *tokenizer) readEscapedCharacter() (rune, error) {
	t.pos++

	if t.pos > len(t.input)-1 {
		return 0, ErrUnexpectedEOF
	}

	switch t.input[t.pos] {
	case '"', '/', '\\':
		t.pos++
		return rune(t.input[t.pos]), nil
	case 'b':
		t.pos++
		return '\b', nil
	case 'f':
		t.pos++
		return '\f', nil
	case 'n':
		t.pos++
		return '\n', nil
	case 'r':
		t.pos++
		return '\r', nil
	case 't':
		t.pos++
		return '\t', nil
	case 'u':
		if t.pos+5 > len(t.input) {
			return 0, ErrUnexpectedEOF
		}

		charValue, err := strconv.ParseInt(t.input[t.pos+1:t.pos+5], 16, 32)
		if err != nil {
			return 0, NewInvalidCharacterError(t.pos, t.input[t.pos+1:t.pos+5])
		}

		unicodeChar := rune(charValue)

		t.pos += 5
		return unicodeChar, nil
	default:
		return 0, NewInvalidCharacterError(t.pos, t.input[t.pos])
	}
}

// byteSet represents a minimal set implementation using a map.
type byteSet map[byte]struct{}

// newByteSet creates a new set with prepopulated elements.
func newByteSet(elements ...byte) byteSet {
	s := make(byteSet)
	for _, element := range elements {
		s[element] = struct{}{}
	}
	return s
}

// has checks if the set contains a given element.
func (s byteSet) has(element byte) bool {
	_, exists := s[element]
	return exists
}
