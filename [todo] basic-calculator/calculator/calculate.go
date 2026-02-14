package calculator

import "errors"

func Calculate(s string) (int, error) {
	return 0, nil
}

func tokenize(s string) ([]token, error) {
	tokenizer := newTokenizer(s)
	return tokenizer.tokenize()
}

type tokenizer struct {
	cur int
	s   string
}

func newTokenizer(s string) *tokenizer {
	return &tokenizer{
		s: s,
	}
}

func (t *tokenizer) tokenize() ([]token, error) {
	if len(t.s) == 0 {
		return nil, errors.New("empty string")
	}

	var tokens []token

	for t.cur < len(t.s) {
		if t.s[t.cur] == ' ' {
			t.cur++
			continue
		}

		if isNumber(t.s[t.cur]) {
			tokens = append(tokens, t.readNumber())
			continue
		}

		if t.s[t.cur] == '+' {
			tokens = append(tokens, token{
				typ: tokenTypePlus,
			})
			t.cur++
			continue
		}

		if t.s[t.cur] == '-' {
			tokens = append(tokens, token{
				typ: tokenTypeMinus,
			})
			t.cur++
			continue
		}

		return nil, errors.New("invalid token")
	}
	return tokens, nil
}

func (t *tokenizer) readNumber() token {
	var val []byte
	for t.cur < len(t.s) && isNumber(t.s[t.cur]) {
		val = append(val, t.s[t.cur])
		t.cur++
	}

	return token{
		typ:   tokenTypeNumber,
		value: string(val),
	}
}

func isNumber(b byte) bool {
	return b >= '0' && b <= '9'
}

type token struct {
	typ   tokenType
	value string
}

type tokenType int

const (
	tokenTypeNumber tokenType = iota
	tokenTypePlus
	tokenTypeMinus
)
