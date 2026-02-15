package calculator

import (
	"errors"
	"fmt"
	"strconv"
)

func Calculate(s string) (int, error) {
	tokens, err := tokenize(s)
	if err != nil {
		return 0, err
	}

	expr, err := parse(tokens)
	if err != nil {
		return 0, err
	}

	return expr.value(), nil
}

func parse(tokens []token) (expression, error) {
	parser := newParser(tokens)
	return parser.parse()
}

type parser struct {
	cur    int
	tokens []token
}

func newParser(tokens []token) *parser {
	return &parser{
		tokens: tokens,
	}
}

func (p *parser) parse() (expression, error) {
	if len(p.tokens) == 0 {
		return nil, errors.New("error empty token list")
	}

	expr, err := newNumberExpr(p.tokens[0])
	if err != nil {
		return nil, err
	}
	p.cur = 1

	for p.cur < len(p.tokens) {
		if p.tokens[p.cur].typ == tokenTypePlus {
			expr, err = p.readPlusExpr(expr)
			if err != nil {
				return nil, err
			}
			continue
		}

		return nil, fmt.Errorf("error unexpected token: %+v", p.tokens[p.cur])
	}

	return expr, nil
}

func (p *parser) readPlusExpr(curExpr expression) (expression, error) {
	p.cur++

	if p.cur >= len(p.tokens) || p.tokens[p.cur].typ != tokenTypeNumber {
		return nil, errors.New("error number token expected")
	}

	rightExpr, err := newNumberExpr(p.tokens[p.cur])
	if err != nil {
		return nil, err
	}
	p.cur++

	return &plusExpr{
		left:  curExpr,
		right: rightExpr,
	}, nil
}

type expression interface {
	value() int
}

type numberExpr struct {
	val int
}

func newNumberExpr(numberToken token) (expression, error) {
	if numberToken.typ != tokenTypeNumber {
		return nil, errors.New("error invalid token for creating number expr")
	}

	val, err := strconv.Atoi(numberToken.value)
	if err != nil {
		return nil, fmt.Errorf("error invalid token value '%s': %v", numberToken.value, err)
	}

	return &numberExpr{
		val: val,
	}, nil
}

func (ne *numberExpr) value() int {
	return ne.val
}

type plusExpr struct {
	left  expression
	right expression
}

func (pe *plusExpr) value() int {
	return pe.left.value() + pe.right.value()
}

// tokenize converts the given expression string into a list of token.
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
