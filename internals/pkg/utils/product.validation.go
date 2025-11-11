package utils

import (
	"github.com/go-playground/validator/v10"
)

func ErrorProductdMsg(fe validator.FieldError) string {
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
	case "min":
		switch field {
		case "Size":
			return "size must have at least 1 item"
		case "Variant":
			return "variant must have at least 1 item"
		default:
			return field + " must have at least " + fe.Param() + " item(s)"
		}
	case "max":
		switch field {
		case "Size":
			return "size can have at most 3 items"
		case "Variant":
			return "variant can have at most 2 items"
		default:
			return field + " can have at most " + fe.Param() + " item(s)"
		}
	case "gt":
		return field + " values must be greater than 0"
	default:
		return field + " is invalid"
	}
}
