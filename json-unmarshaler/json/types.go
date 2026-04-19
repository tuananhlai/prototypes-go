package json

import (
	"errors"
	"fmt"
)

type UnexpectedTokenError[T byte | string] struct {
	pos   int
	token T
}

func NewUnexpectedTokenError[T byte | string](pos int, token T) *UnexpectedTokenError[T] {
	return &UnexpectedTokenError[T]{
		pos:   pos,
		token: token,
	}
}

func (u *UnexpectedTokenError[T]) Error() string {
	return fmt.Sprintf("unexpected token '%v' at position %v", string(u.token), u.pos)
}

type InvalidCharacterError[T byte | string] struct {
	pos  int
	char T
}

func NewInvalidCharacterError[T byte | string](pos int, char T) *InvalidCharacterError[T] {
	return &InvalidCharacterError[T]{
		pos:  pos,
		char: char,
	}
}

func (i *InvalidCharacterError[T]) Error() string {
	return fmt.Sprintf("invalid character '%v' found at position %v", string(i.char), i.pos)
}

type InvalidNumberError struct {
	pos    int
	detail string
}

func NewInvalidNumberError(pos int, detail string) *InvalidNumberError {
	return &InvalidNumberError{
		pos:    pos,
		detail: detail,
	}
}

func (i *InvalidNumberError) Error() string {
	return fmt.Sprintf("invalid number found at position %v: %v", i.pos, i.detail)
}

var (
	ErrUnexpectedEOF          = errors.New("unexpected end of JSON input")
	ErrUnknownState           = errors.New("error unknown state reached")
	ErrInvalidRootObject      = errors.New("invalid root object")
	ErrMultipleRootElements   = errors.New("multiple root elements found")
	ErrGenericUnexpectedToken = errors.New("unexpected token")
)
