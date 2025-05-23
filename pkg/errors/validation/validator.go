package validation

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/errors/types"
)

// Validator interface for different validation strategies
type Validator interface {
	Validate(ctx context.Context, value interface{}) error
}

// ValidationRule represents a single validation rule
type ValidationRule interface {
	Apply(value interface{}) error
	Name() string
}

// ValidationError represents validation errors with details
type ValidationError struct {
	Field   string   `json:"field"`
	Code    string   `json:"code"`
	Message string   `json:"message"`
	Value   interface{} `json:"value,omitempty"`
}

// Error implements the error interface
func (ve *ValidationError) Error() string {
	return fmt.Sprintf("validation failed for field '%s': %s", ve.Field, ve.Message)
}

// MultiValidationError represents multiple validation errors
type MultiValidationError struct {
	Errors []ValidationError `json:"errors"`
}

// Error implements the error interface
func (mve *MultiValidationError) Error() string {
	if len(mve.Errors) == 0 {
		return "validation failed"
	}
	
	if len(mve.Errors) == 1 {
		return mve.Errors[0].Error()
	}
	
	var messages []string
	for _, err := range mve.Errors {
		messages = append(messages, err.Error())
	}
	
	return fmt.Sprintf("multiple validation errors: %s", strings.Join(messages, "; "))
}

// RequiredRule validates that a value is not empty
type RequiredRule struct {
	FieldName string
}

func (r *RequiredRule) Apply(value interface{}) error {
	if value == nil {
		return &ValidationError{
			Field:   r.FieldName,
			Code:    types.ErrMissingRequired,
			Message: "field is required",
			Value:   value,
		}
	}
	
	switch v := value.(type) {
	case string:
		if strings.TrimSpace(v) == "" {
			return &ValidationError{
				Field:   r.FieldName,
				Code:    types.ErrMissingRequired,
				Message: "field cannot be empty",
				Value:   value,
			}
		}
	case []interface{}:
		if len(v) == 0 {
			return &ValidationError{
				Field:   r.FieldName,
				Code:    types.ErrMissingRequired,
				Message: "field cannot be empty",
				Value:   value,
			}
		}
	}
	
	return nil
}

func (r *RequiredRule) Name() string {
	return "required"
}

// FormatRule validates that a value matches a specific format
type FormatRule struct {
	FieldName string
	Pattern   *regexp.Regexp
	Format    string
}

func (r *FormatRule) Apply(value interface{}) error {
	if value == nil {
		return nil // Skip format validation for nil values
	}
	
	str, ok := value.(string)
	if !ok {
		return &ValidationError{
			Field:   r.FieldName,
			Code:    types.ErrInvalidFormat,
			Message: "field must be a string",
			Value:   value,
		}
	}
	
	if !r.Pattern.MatchString(str) {
		return &ValidationError{
			Field:   r.FieldName,
			Code:    types.ErrInvalidFormat,
			Message: fmt.Sprintf("field must match format: %s", r.Format),
			Value:   value,
		}
	}
	
	return nil
}

func (r *FormatRule) Name() string {
	return "format"
}

// RangeRule validates that a numeric value is within a specific range
type RangeRule struct {
	FieldName string
	Min       float64
	Max       float64
}

func (r *RangeRule) Apply(value interface{}) error {
	if value == nil {
		return nil // Skip range validation for nil values
	}
	
	var num float64
	var ok bool
	
	switch v := value.(type) {
	case int:
		num = float64(v)
		ok = true
	case int32:
		num = float64(v)
		ok = true
	case int64:
		num = float64(v)
		ok = true
	case float32:
		num = float64(v)
		ok = true
	case float64:
		num = v
		ok = true
	}
	
	if !ok {
		return &ValidationError{
			Field:   r.FieldName,
			Code:    types.ErrInvalidFormat,
			Message: "field must be a number",
			Value:   value,
		}
	}
	
	if num < r.Min || num > r.Max {
		return &ValidationError{
			Field:   r.FieldName,
			Code:    types.ErrOutOfRange,
			Message: fmt.Sprintf("field must be between %g and %g", r.Min, r.Max),
			Value:   value,
		}
	}
	
	return nil
}

func (r *RangeRule) Name() string {
	return "range"
}

// FieldValidator validates a specific field with multiple rules
type FieldValidator struct {
	FieldName string
	Rules     []ValidationRule
}

// NewFieldValidator creates a new field validator
func NewFieldValidator(fieldName string) *FieldValidator {
	return &FieldValidator{
		FieldName: fieldName,
		Rules:     make([]ValidationRule, 0),
	}
}

// Required adds a required rule
func (fv *FieldValidator) Required() *FieldValidator {
	fv.Rules = append(fv.Rules, &RequiredRule{FieldName: fv.FieldName})
	return fv
}

// Format adds a format rule
func (fv *FieldValidator) Format(pattern, description string) *FieldValidator {
	regex := regexp.MustCompile(pattern)
	fv.Rules = append(fv.Rules, &FormatRule{
		FieldName: fv.FieldName,
		Pattern:   regex,
		Format:    description,
	})
	return fv
}

// Range adds a range rule
func (fv *FieldValidator) Range(min, max float64) *FieldValidator {
	fv.Rules = append(fv.Rules, &RangeRule{
		FieldName: fv.FieldName,
		Min:       min,
		Max:       max,
	})
	return fv
}

// Validate validates a value against all rules
func (fv *FieldValidator) Validate(value interface{}) error {
	var errors []ValidationError
	
	for _, rule := range fv.Rules {
		if err := rule.Apply(value); err != nil {
			if ve, ok := err.(*ValidationError); ok {
				errors = append(errors, *ve)
			}
		}
	}
	
	if len(errors) > 0 {
		return &MultiValidationError{Errors: errors}
	}
	
	return nil
}

// StructValidator validates entire structures
type StructValidator struct {
	validators map[string]*FieldValidator
}

// NewStructValidator creates a new struct validator
func NewStructValidator() *StructValidator {
	return &StructValidator{
		validators: make(map[string]*FieldValidator),
	}
}

// Field adds a field validator
func (sv *StructValidator) Field(name string) *FieldValidator {
	validator := NewFieldValidator(name)
	sv.validators[name] = validator
	return validator
}

// Validate validates a map of values
func (sv *StructValidator) Validate(ctx context.Context, values map[string]interface{}) error {
	var allErrors []ValidationError
	
	for fieldName, validator := range sv.validators {
		value, exists := values[fieldName]
		if !exists {
			value = nil
		}
		
		if err := validator.Validate(value); err != nil {
			if mve, ok := err.(*MultiValidationError); ok {
				allErrors = append(allErrors, mve.Errors...)
			} else if ve, ok := err.(*ValidationError); ok {
				allErrors = append(allErrors, *ve)
			}
		}
	}
	
	if len(allErrors) > 0 {
		return types.NewAppError(
			types.ErrInvalidInput,
			"input validation failed",
			&MultiValidationError{Errors: allErrors},
		)
	}
	
	return nil
}

// Common validation patterns
var (
	EmailPattern    = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	URLPattern      = `^https?://[^\s/$.?#].[^\s]*$`
	APIKeyPattern   = `^[a-zA-Z0-9]{32,}$`
	TokenPattern    = `^[a-zA-Z0-9._-]+$`
	FilenamePattern = `^[a-zA-Z0-9._-]+$`
) 