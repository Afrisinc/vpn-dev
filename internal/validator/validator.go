package validator

import (
	"fmt"
	"net"
	"regexp"

	"github.com/go-playground/validator/v10"
)

// Validator wraps the validator library with custom validation rules
type Validator struct {
	v *validator.Validate
}

// ValidationError represents a field validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrors is a collection of field validation errors
type ValidationErrors []ValidationError

// New creates a new validator instance with custom rules
func New() *Validator {
	v := validator.New()

	// Register custom validators
	v.RegisterValidation("wireguard_key", validateWireGuardKey)
	v.RegisterValidation("ipv4", validateIPv4)

	return &Validator{v: v}
}

// Validate validates a struct and returns field-level errors
func (v *Validator) Validate(data interface{}) ValidationErrors {
	err := v.v.Struct(data)
	if err == nil {
		return nil
	}

	var errors ValidationErrors
	validationErrs, ok := err.(validator.ValidationErrors)
	if !ok {
		return ValidationErrors{
			{
				Field:   "unknown",
				Message: err.Error(),
			},
		}
	}

	for _, fieldErr := range validationErrs {
		errors = append(errors, ValidationError{
			Field:   fieldErr.Field(),
			Message: formatValidationError(fieldErr),
		})
	}

	return errors
}

// formatValidationError formats a single field validation error
func formatValidationError(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", err.Field())
	case "email":
		return fmt.Sprintf("%s must be a valid email address", err.Field())
	case "ipv4":
		return fmt.Sprintf("%s must be a valid IPv4 address", err.Field())
	case "wireguard_key":
		return fmt.Sprintf("%s must be a valid WireGuard public key", err.Field())
	case "oneof":
		return fmt.Sprintf("%s must be one of the allowed values", err.Field())
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", err.Field())
	case "ip":
		return fmt.Sprintf("%s must be a valid IP address", err.Field())
	default:
		return fmt.Sprintf("%s failed validation: %s", err.Field(), err.Tag())
	}
}

// validateWireGuardKey validates WireGuard public key format
// WireGuard keys are base64-encoded 32-byte values, resulting in 44 characters
func validateWireGuardKey(fl validator.FieldLevel) bool {
	key := fl.Field().String()

	// Check length (44 characters for base64-encoded 32-byte key)
	if len(key) != 44 {
		return false
	}

	// Check if it's valid base64
	base64Regex := regexp.MustCompile(`^[A-Za-z0-9+/]*={0,2}$`)
	if !base64Regex.MatchString(key) {
		return false
	}

	return true
}

// validateIPv4 validates that a string is a valid IPv4 address
func validateIPv4(fl validator.FieldLevel) bool {
	ip := fl.Field().String()
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return false
	}
	// Ensure it's IPv4, not IPv6
	return parsed.To4() != nil
}

// validateOneof is a custom validator for enum-like values
// Note: This overrides the default oneof to provide better flexibility
func validateOneof(fl validator.FieldLevel) bool {
	// The actual validation is handled by the default validator
	// This is here as a placeholder for custom logic if needed
	return true
}
