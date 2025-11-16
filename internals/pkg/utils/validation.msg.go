package utils

import (
	"errors"

	"github.com/go-playground/validator/v10"
)

func ErrorMessage(fe validator.FieldError) string {
	field := fe.Field()
	tag := fe.Tag()

	// --- Custom messages per field + tag (Product) ---
	switch field {
	case "Rating":
		if tag == "gte" || tag == "lte" {
			return "rating must be between 1 - 10"
		}
	case "Price":
		if tag == "gte" {
			return "price must be at least 5000"
		}
	case "Stock":
		if tag == "gte" {
			return "stock must be 0 or more"
		}
	case "Size":
		if tag == "min" {
			return "size must have at least 1 item"
		}
		if tag == "max" {
			return "size can have at most 3 items"
		}
	case "Variant":
		if tag == "min" {
			return "variant must have at least 1 item"
		}
		if tag == "max" {
			return "variant can have at most 2 items"
		}
	}

	// --- Custom messages per field + tag (Register/Login/User) ---
	switch field {
	case "Email":
		if tag == "email" {
			return "invalid email format"
		}
	case "Username":
		if tag == "max" {
			return field + " must be at most " + fe.Param() + " characters"
		}
	case "Role":
		if tag == "oneof" {
			return field + " must be either 'user' or 'admin'"
		}
	}

	// --- Custom messages per field + tag (Phone/User) ---
	switch field {
	case "Phone":
		if tag == "len" {
			return "phone must be exactly 12 digits"
		}
		if tag == "numeric" {
			return "phone must contain only numbers"
		}
	}

	// --- Default messages ---
	switch tag {
	case "password_complex":
		return "password must contain uppercase, lowercase, number, and special character"
	case "required":
		return field + " is required"
	case "gt":
		return field + " must be greater than " + fe.Param()
	case "gte":
		return field + " must be greater than or equal to " + fe.Param()
	case "lte":
		return field + " must be less than or equal to " + fe.Param()
	case "lt":
		return field + " must be less than " + fe.Param()
	case "min":
		return field + " must have at least " + fe.Param() + " item(s)"
	case "max":
		return field + " can have at most " + fe.Param() + " item(s)"
	case "eqfield":
		return field + " must match " + fe.Param()
	default:
		return field + " is invalid"
	}
}

// --- ERROR MESSAGE UPDATE PASSWORD ---
var ErrValidation = errors.New("validation error")
