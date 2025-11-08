package utils

import (
	"github.com/go-playground/validator/v10"
)

func ErrorRegisterMsg(fe validator.FieldError) string {
	field := fe.Field()
	switch fe.Tag() {
	case "required":
		return field + " is required"
	case "max":
		return field + " must be at most " + fe.Param() + " characters"
	case "password_complex":
		return field + " must contain uppercase, lowercase, number, and special character"
	case "email":
		return "invalid email format"
	default:
		return field + " is invalid"
	}
}

func ErrorLoginMsg(fe validator.FieldError) string {
	field := fe.Field()
	switch fe.Tag() {
	case "required":
		return field + " is required"
	case "password_complex":
		return field + " must contain uppercase, lowercase, number, and special character"
	case "email":
		return "invalid email format"
	default:
		return field + " is invalid"
	}
}
