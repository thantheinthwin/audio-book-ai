package utils

import (
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

// GetValidator returns the global validator instance
func GetValidator() *validator.Validate {
	return validate
}
