package middleware

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// Validator wraps the validator instance
type Validator struct {
	validate *validator.Validate
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Value   string `json:"value,omitempty"`
	Message string `json:"message"`
}

// ValidationErrorResponse is the response for validation errors
type ValidationErrorResponse struct {
	Success bool              `json:"success"`
	Message string            `json:"message"`
	Errors  []ValidationError `json:"errors"`
}

// NewValidator creates a new validator instance
func NewValidator() *Validator {
	v := validator.New()

	// Register function to get json tag name
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return &Validator{validate: v}
}

// Validate validates a struct
func (v *Validator) Validate(i interface{}) []ValidationError {
	var errors []ValidationError

	if err := v.validate.Struct(i); err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			errors = append(errors, ValidationError{
				Field:   err.Field(),
				Tag:     err.Tag(),
				Value:   err.Param(),
				Message: getErrorMessage(err),
			})
		}
	}

	return errors
}

// RegisterValidation registers a custom validation
func (v *Validator) RegisterValidation(tag string, fn validator.Func) error {
	return v.validate.RegisterValidation(tag, fn)
}

// ValidateMiddleware returns a Fiber middleware for request validation
func ValidateMiddleware[T any](v *Validator) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var body T

		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(ValidationErrorResponse{
				Success: false,
				Message: "Invalid request body",
				Errors: []ValidationError{
					{Field: "body", Tag: "parse", Message: "Failed to parse request body"},
				},
			})
		}

		errors := v.Validate(body)
		if len(errors) > 0 {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(ValidationErrorResponse{
				Success: false,
				Message: "Validation failed",
				Errors:  errors,
			})
		}

		// Store validated body in context
		c.Locals("body", body)
		return c.Next()
	}
}

// ValidateBody validates the request body and returns errors
func ValidateBody[T any](c *fiber.Ctx, v *Validator) (*T, *ValidationErrorResponse) {
	var body T

	if err := c.BodyParser(&body); err != nil {
		return nil, &ValidationErrorResponse{
			Success: false,
			Message: "Invalid request body",
			Errors: []ValidationError{
				{Field: "body", Tag: "parse", Message: "Failed to parse request body"},
			},
		}
	}

	errors := v.Validate(body)
	if len(errors) > 0 {
		return nil, &ValidationErrorResponse{
			Success: false,
			Message: "Validation failed",
			Errors:  errors,
		}
	}

	return &body, nil
}

// ValidateQuery validates query parameters
func ValidateQuery[T any](c *fiber.Ctx, v *Validator) (*T, *ValidationErrorResponse) {
	var query T

	if err := c.QueryParser(&query); err != nil {
		return nil, &ValidationErrorResponse{
			Success: false,
			Message: "Invalid query parameters",
			Errors: []ValidationError{
				{Field: "query", Tag: "parse", Message: "Failed to parse query parameters"},
			},
		}
	}

	errors := v.Validate(query)
	if len(errors) > 0 {
		return nil, &ValidationErrorResponse{
			Success: false,
			Message: "Validation failed",
			Errors:  errors,
		}
	}

	return &query, nil
}

// getErrorMessage returns a human-readable error message
func getErrorMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email format"
	case "min":
		return "Value is too short or too small (minimum: " + fe.Param() + ")"
	case "max":
		return "Value is too long or too large (maximum: " + fe.Param() + ")"
	case "len":
		return "Value must be exactly " + fe.Param() + " characters"
	case "gte":
		return "Value must be greater than or equal to " + fe.Param()
	case "lte":
		return "Value must be less than or equal to " + fe.Param()
	case "gt":
		return "Value must be greater than " + fe.Param()
	case "lt":
		return "Value must be less than " + fe.Param()
	case "eqfield":
		return "Value must match " + fe.Param()
	case "nefield":
		return "Value must not match " + fe.Param()
	case "oneof":
		return "Value must be one of: " + fe.Param()
	case "url":
		return "Invalid URL format"
	case "uuid":
		return "Invalid UUID format"
	case "alpha":
		return "Value must contain only alphabetic characters"
	case "alphanum":
		return "Value must contain only alphanumeric characters"
	case "numeric":
		return "Value must be numeric"
	case "mobile":
		return "Invalid mobile number format"
	case "password":
		return "Password must contain at least one uppercase, one lowercase, one number, and one special character"
	default:
		return "Invalid value"
	}
}

// Helper function to get validated body from context
func GetValidatedBody[T any](c *fiber.Ctx) (T, bool) {
	body, ok := c.Locals("body").(T)
	return body, ok
}
