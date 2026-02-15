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
		if p.tokens[p.cur].typ == tokenTypePlus || p.tokens[p.cur].typ == tokenTypeMinus {
			expr, err = p.readBinaryExpr(expr)
			if err != nil {
				return nil, err
			}
			continue
		}

		return nil, fmt.Errorf("error unexpected token: %+v", p.tokens[p.cur])
	}

	return expr, nil
}

// readBinaryExpr creates a binary expression by consuming an operator token and the right operand.
// This function must be called when the cursor is on an operator token.
func (p *parser) readBinaryExpr(leftOperand expression) (expression, error) {
	opToken := p.tokens[p.cur]
	p.cur++

	if p.cur >= len(p.tokens) {
		return nil, errors.New("error unexpected end of expression")
	}

	var err error
	var rightOperand expression
	switch p.tokens[p.cur].typ {
	case tokenTypeMinus:
		rightOperand, err = p.readNegativeExpr()
	case tokenTypeNumber:
		rightOperand, err = p.readNumber()
	default:
		return nil, fmt.Errorf("error unexpected token: %+v", p.tokens[p.cur])
	}
	if err != nil {
		return nil, err
	}

	return newBinaryExpr(leftOperand, rightOperand, opToken)
}

func (p *parser) readNumber() (expression, error) {
	numberExpr, err := newNumberExpr(p.tokens[p.cur])
	if err != nil {
		return nil, err
	}
	p.cur++

	return numberExpr, nil
}

// readNegativeExpr ...
func (p *parser) readNegativeExpr() (expression, error) {
	p.cur++

	if p.cur >= len(p.tokens) {
		return nil, errors.New("error unexpected end of expression")
	}

	if p.tokens[p.cur].typ != tokenTypeNumber {
		return nil, fmt.Errorf("error unexpected token encounter: %+v", p.tokens[p.cur])
	}

	numberExpr, err := newNumberExpr(p.tokens[p.cur])
	if err != nil {
		return nil, err
	}
	p.cur++

	return &negativeExpr{
		operand: numberExpr,
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

type negativeExpr struct {
	operand expression
}

func (ne *negativeExpr) value() int {
	return -ne.operand.value()
}

type binaryExpr struct {
	left    expression
	right   expression
	opToken token
}

func newBinaryExpr(left expression, right expression, opToken token) (expression, error) {
	if opToken.typ != tokenTypePlus && opToken.typ != tokenTypeMinus {
		return nil, fmt.Errorf("error invalid operator token: got %+v", opToken)
	}

	return &binaryExpr{
		left:    left,
		right:   right,
		opToken: opToken,
	}, nil
}

func (pe *binaryExpr) value() int {
	if pe.opToken.typ == tokenTypePlus {
		return pe.left.value() + pe.right.value()
	}
	if pe.opToken.typ == tokenTypeMinus {
		return pe.left.value() - pe.right.value()
	}
	panic("unknown state reached. invalid operator token")
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

// readNumber consumes all numeric digits while advancing the cursor at the same time
// to create a number token. It must be called when the cursor is on a numeric character.
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
