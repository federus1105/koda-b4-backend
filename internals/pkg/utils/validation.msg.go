package utils

import "github.com/go-playground/validator/v10"

func ErrorMessage(fe validator.FieldError) string {
	field := fe.Field()
	tag := fe.Tag()  

	switch tag {
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

	default:
		return field + " is invalid"
	}
}
