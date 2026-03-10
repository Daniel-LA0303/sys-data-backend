package validator

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func ValidateStruct(input interface{}) map[string]string {

	err := validate.Struct(input)

	if err == nil {
		return nil
	}

	errors := make(map[string]string)

	for _, err := range err.(validator.ValidationErrors) {
		field := err.Field()

		switch err.Tag() {
		case "required":
			errors[field] = fmt.Sprintf("%s is required", field)
		case "email":
			errors[field] = "invalid email format"
		case "min":
			errors[field] = fmt.Sprintf("%s must have minimum length %s", field, err.Param())
		default:
			errors[field] = "invalid value"
		}
	}

	return errors
}
