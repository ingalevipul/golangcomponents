package main

import (
	"github.com/go-playground/validator/v10"
)

var Validator = validator.New()

func Validate(s any) error {
	return Validator.Struct(s)
}
