package cheatsheettest

import (
	"net/mail"
	"regexp"

	"github.com/go-playground/validator/v10"
)

func ValidateEmailRegex(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9.!#$%&'*+/=?^_` + "`" +
		`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)
	return emailRegex.MatchString(email)
}

func ValidateEmailStd(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func ValidateEmailValidator(email string) bool {
	validate := validator.New()
	return validate.Var(email, "email") == nil
}
