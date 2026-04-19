package json

import "fmt"

// Parse parses a JSON string into a Go data structure.
func Parse(input string) (interface{}, error) {
	tokenizer := newTokenizer(input)
	tokens, err := tokenizer.tokenize()
	if err != nil {
		return nil, fmt.Errorf("error tokenizing input: %w", err)
	}

	parser := newParser(tokens)
	output, err := parser.parse()
	if err != nil {
		return nil, fmt.Errorf("error parsing input: %w", err)
	}

	return output, nil
}
