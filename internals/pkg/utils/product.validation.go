package utils

import (
	"github.com/go-playground/validator/v10"
)

func ErrorCreatedMsg(fe validator.FieldError) string {
	field := fe.Field()
	switch fe.Tag() {
	case "required":
		return field + " is required"
	case "gte":
		switch field {
		case "Rating":
			return "rating must be between 1 - 10"
		case "Price":
			return "price must be at least 5000"
		case "Stock":
			return "stock must be 0 or more"
		default:
			return field + " must be greater than or equal to " + fe.Param()
		}
	case "lte":
		switch field {
		case "Rating":
			return "rating must be between 1 - 10"
		default:
			return field + " must be less than or equal to " + fe.Param()
		}
	default:
		return field + " is invalid"
	}
}
