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
	tokens = append([]token{{typ: tokenTypeLParen}}, tokens...)
	tokens = append(tokens, token{typ: tokenTypeRParen})
	return &parser{
		tokens: tokens,
	}
}

func (p *parser) parse() (expression, error) {
	if len(p.tokens) == 0 {
		return nil, errors.New("error empty token list")
	}

	return p.readPrimaryExpr()
}

func (p *parser) readPrimaryExpr() (expression, error) {
	if p.cur >= len(p.tokens) {
		return nil, errors.New("error unexpected end of expression")
	}
	var expr expression
	var err error

	curToken := p.tokens[p.cur]
	switch curToken.typ {
	case tokenTypeNumber:
		expr, err = p.readNumber()
	case tokenTypeLParen:
		expr, err = p.readParenExpr()
	case tokenTypeMinus:
		expr, err = p.readNegativeExpr()
	default:
		return nil, fmt.Errorf("error unexpected token: %+v", curToken)
	}

	if err != nil {
		return nil, err
	}

	return expr, err
}

func (p *parser) readParenExpr() (expression, error) {
	if p.cur >= len(p.tokens) || p.tokens[p.cur].typ != tokenTypeLParen {
		return nil, errors.New("error unexpected token or eof")
	}
	p.cur++

	var expr expression
	var err error

	for p.cur < len(p.tokens) {
		curToken := p.tokens[p.cur]

		if expr == nil {
			expr, err = p.readPrimaryExpr()
		} else if curToken.typ == tokenTypePlus || curToken.typ == tokenTypeMinus {
			expr, err = p.readBinaryExpr(expr)
		} else if curToken.typ == tokenTypeRParen {
			p.cur++
			return expr, nil
		} else {
			return nil, fmt.Errorf("error unexpected token: %+v", curToken)
		}

		if err != nil {
			return nil, err
		}
	}

	return nil, errors.New("error unexpected end of expression")
}

// readBinaryExpr creates a binary expression by consuming an operator token and the right operand.
// This function must be called when the cursor is on an operator token.
func (p *parser) readBinaryExpr(leftOperand expression) (expression, error) {
	if p.cur >= len(p.tokens) || (p.tokens[p.cur].typ != tokenTypeMinus && p.tokens[p.cur].typ != tokenTypePlus) {
		return nil, errors.New("error unexpected end of expression")
	}
	opToken := p.tokens[p.cur]
	p.cur++

	rightOperand, err := p.readPrimaryExpr()
	if err != nil {
		return nil, err
	}

	return newBinaryExpr(leftOperand, rightOperand, opToken)
}

func (p *parser) readNumber() (expression, error) {
	if p.cur >= len(p.tokens) || p.tokens[p.cur].typ != tokenTypeNumber {
		return nil, errors.New("error unexpected end of expression")
	}

	numberExpr, err := newNumberExpr(p.tokens[p.cur])
	if err != nil {
		return nil, err
	}
	p.cur++

	return numberExpr, nil
}

// readNegativeExpr ...
func (p *parser) readNegativeExpr() (expression, error) {
	if p.cur >= len(p.tokens) || p.tokens[p.cur].typ != tokenTypeMinus {
		return nil, errors.New("error unexpected end of expression")
	}
	p.cur++

	operand, err := p.readPrimaryExpr()
	if err != nil {
		return nil, err
	}

	return &negativeExpr{
		operand: operand,
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
	tokenTypeLParen
	tokenTypeRParen
)
