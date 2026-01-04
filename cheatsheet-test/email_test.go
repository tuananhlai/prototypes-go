package cheatsheettest_test

import (
	"fmt"
	"testing"

	cheatsheettest "github.com/tuananhlai/prototypes/cheatsheet-test"
)

func TestValidateEmailRegex(t *testing.T) {
	testValidateEmail(t, cheatsheettest.ValidateEmailRegex)
}

func TestValidateEmailStd(t *testing.T) {
	testValidateEmail(t, cheatsheettest.ValidateEmailStd)
}

func TestValidateEmailValidator(t *testing.T) {
	testValidateEmail(t, cheatsheettest.ValidateEmailValidator)
}

func testValidateEmail(t *testing.T, fn func(email string) bool) {
	tests := []struct {
		arg  string
		want bool
	}{
		{arg: "test@example.com", want: true},
		{arg: "test@example.com.br", want: true},
		{arg: "test+extra@example.com.br", want: true},
		{arg: "test", want: false},
		{arg: "test@example", want: false},
		{arg: "test@example..com", want: false},
		{arg: "test@example.com.", want: false},
		{arg: "test space@example.com", want: false},
		{arg: "test@example.com.", want: false},
		{arg: "test😊@example.com", want: false},
		{arg: "-test@example.com", want: true},
	}

	for _, tc := range tests {
		validMsg := "valid"
		if !tc.want {
			validMsg = "invalid"
		}
		t.Run(fmt.Sprintf("%q should be %s", tc.arg, validMsg), func(tt *testing.T) {
			assertEqual(tt, fn(tc.arg), tc.want)
		})
	}
}

func assertEqual(t *testing.T, got, want bool) {
	// Mark the function as a test helper, so that Go test will report the line number of the call site, not the
	// `assertEqual` function.
	t.Helper()
	if got != want {
		t.Errorf("got %t, want %t", got, want)
	}
}
