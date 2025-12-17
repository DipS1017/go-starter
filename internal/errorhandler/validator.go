package errorhandler

import (
	"github.com/go-playground/validator/v10"
)

type CustomValidator struct {
	validator *validator.Validate
}

// NewValidator initializes the custom validator
func NewValidator() *CustomValidator {
	return &CustomValidator{validator: validator.New()}
}

// Validate implements the echo.Validator interface
func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return ErrorBadRequest(TranslateValidationErrors(err))
	}
	return nil
}

func TranslateValidationErrors(err error) map[string]string {
	errors := make(map[string]string)
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fieldErr := range validationErrors {
			fieldName := fieldErr.Field()
			if fieldErr.Tag() == "required" {
				errors[fieldName] = fieldName + " is required"
			} else if fieldErr.Tag() == "email" {
				errors[fieldName] = fieldName + " must be a valid email address"
			} else if fieldErr.Tag() == "min" {
				errors[fieldName] = fieldName + " must be at least " + fieldErr.Param() + " characters long"
			} else {
				errors[fieldName] = fieldName + " is invalid"
			}
		}
	} else {
		errors["error"] = "Invalid input data: " + err.Error()
	}
	return errors
}
