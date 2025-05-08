package validator

import (
	"errors"
	"log"
	"regexp"
	"strings"
)

type Validator struct{}

func NewValidator() *Validator {
	return &Validator{}
}

// simple regex pattern that matches most valid email formats
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func (v *Validator) ValidateCreateUser(firstName, lastName, email string) error {
	if strings.TrimSpace(firstName) == "" || strings.TrimSpace(lastName) == "" {
		return errors.New("firstName or lastName is empty")
	}
	log.Println(firstName, email)
	if !emailRegex.MatchString(email) {
		return errors.New("email is invalid")
	}
	return nil
}
