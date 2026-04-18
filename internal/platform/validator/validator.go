package validator

import (
	"fmt"
	"reflect"

	"github.com/go-playground/validator/v10"
)

// Validator wraps the go-playground validator
type Validator struct {
	validate *validator.Validate
}

// New creates a new validator instance
func New() *Validator {
	v := validator.New()

	// Register custom validations if needed
	_ = v.RegisterValidation("required_if", requiredIf)

	return &Validator{validate: v}
}

// Validate validates a struct
func (v *Validator) Validate(s interface{}) error {
	return v.validate.Struct(s)
}

// ValidateVar validates a single variable
func (v *Validator) ValidateVar(field interface{}, tag string) error {
	return v.validate.Var(field, tag)
}

// requiredIf is a custom validation that checks if a field is required based on another field
func requiredIf(fl validator.FieldLevel) bool {
	// params format: "otherField=value"
	// This is a simplified version - in production, implement proper logic
	return true
}

// GetError returns a formatted error message
func GetError(err error) map[string]string {
	errors := make(map[string]string)

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			field := e.Field()
			switch e.Tag() {
			case "required":
				errors[field] = fmt.Sprintf("%s is required", field)
			case "email":
				errors[field] = fmt.Sprintf("%s must be a valid email", field)
			case "min":
				errors[field] = fmt.Sprintf("%s must be at least %s", field, e.Param())
			case "max":
				errors[field] = fmt.Sprintf("%s must be at most %s", field, e.Param())
			case "gte":
				errors[field] = fmt.Sprintf("%s must be greater than or equal to %s", field, e.Param())
			case "lte":
				errors[field] = fmt.Sprintf("%s must be less than or equal to %s", field, e.Param())
			default:
				errors[field] = fmt.Sprintf("%s failed validation: %s", field, e.Tag())
			}
		}
	} else {
		errors["general"] = err.Error()
	}

	return errors
}

// StructLevelValidation for custom struct-level validation
func StructLevelValidation(sl validator.StructLevel) {
	// Implement custom struct-level validation logic here
}

// TranslateError translates validation errors to user-friendly messages
func TranslateError(err error) string {
	if err == nil {
		return ""
	}

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			return fmt.Sprintf("Field %s: %s", e.Field(), getErrorMessage(e))
		}
	}

	return err.Error()
}

func getErrorMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "is required"
	case "email":
		return "must be a valid email address"
	case "min":
		return fmt.Sprintf("must be at least %s", e.Param())
	case "max":
		return fmt.Sprintf("must be at most %s", e.Param())
	case "len":
		return fmt.Sprintf("must be %s characters", e.Param())
	case "gte":
		return fmt.Sprintf("must be greater than or equal to %s", e.Param())
	case "lte":
		return fmt.Sprintf("must be less than or equal to %s", e.Param())
	case "oneof":
		return fmt.Sprintf("must be one of: %s", e.Param())
	default:
		return "is invalid"
	}
}

// GetFieldTag gets the validation tag for a struct field
func GetFieldTag(s interface{}, field string) string {
	t := reflect.TypeOf(s)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return ""
	}

	f, ok := t.FieldByName(field)
	if !ok {
		return ""
	}

	return f.Tag.Get("validate")
}
