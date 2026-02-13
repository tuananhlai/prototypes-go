package calculator

func Calculate(s string) (int, error) {
	return 0, nil
}

type tokenizer struct {
	cur int
	s   string
}

func (t *tokenizer) tokenize() ([]token, error) {
	// if len(s) == 0 {
	// 	return nil, errors.New("empty string")
	// }

	// var tokens []token

	// for t.cur < len(s) {
	// 	if t.s[t.cur] == ' ' {
	// 		t.cur++
	// 		continue
	// 	}
	// }
	return nil, nil
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
