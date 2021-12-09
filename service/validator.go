package service

import "github.com/go-playground/validator/v10"

func NewValidate() *validator.Validate {
	validate := validator.New()
	return validate
}
