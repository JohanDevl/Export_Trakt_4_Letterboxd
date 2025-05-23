package validation

import (
	"context"
	"testing"
)

func TestFieldValidatorRequired(t *testing.T) {
	validator := NewFieldValidator("test_field")
	validator.Required()
	
	// Test empty string
	err := validator.Validate("")
	if err == nil {
		t.Error("Expected validation error for empty string")
	}
	
	// Test nil value
	err = validator.Validate(nil)
	if err == nil {
		t.Error("Expected validation error for nil value")
	}
	
	// Test valid value
	err = validator.Validate("test")
	if err != nil {
		t.Errorf("Expected no error for valid value, got: %v", err)
	}
}

func TestFieldValidatorFormat(t *testing.T) {
	validator := NewFieldValidator("test_field")
	validator.Format(`^[a-zA-Z]+$`, "Must contain only letters")
	
	// Test invalid format
	err := validator.Validate("test123")
	if err == nil {
		t.Error("Expected validation error for invalid format")
	}
	
	// Test valid format
	err = validator.Validate("test")
	if err != nil {
		t.Errorf("Expected no error for valid format, got: %v", err)
	}
}

func TestFieldValidatorRange(t *testing.T) {
	validator := NewFieldValidator("test_field")
	validator.Range(1, 10)
	
	// Test below range
	err := validator.Validate(0)
	if err == nil {
		t.Error("Expected validation error for value below range")
	}
	
	// Test above range
	err = validator.Validate(11)
	if err == nil {
		t.Error("Expected validation error for value above range")
	}
	
	// Test within range
	err = validator.Validate(5)
	if err != nil {
		t.Errorf("Expected no error for value within range, got: %v", err)
	}
	
	// Test edge cases
	err = validator.Validate(1)
	if err != nil {
		t.Errorf("Expected no error for minimum value, got: %v", err)
	}
	
	err = validator.Validate(10)
	if err != nil {
		t.Errorf("Expected no error for maximum value, got: %v", err)
	}
}

func TestFieldValidatorChaining(t *testing.T) {
	validator := NewFieldValidator("test_field")
	validator.Required().Format(`^[a-zA-Z]+$`, "Must contain only letters")
	
	// Test empty string (should fail on required)
	err := validator.Validate("")
	if err == nil {
		t.Error("Expected validation error for empty string")
	}
	
	// Test invalid format
	err = validator.Validate("test123")
	if err == nil {
		t.Error("Expected validation error for invalid format")
	}
	
	// Test valid value
	err = validator.Validate("test")
	if err != nil {
		t.Errorf("Expected no error for valid value, got: %v", err)
	}
}

func TestStructValidator(t *testing.T) {
	validator := NewStructValidator()
	validator.Field("name").Required().Format(`^[a-zA-Z\s]+$`, "Name must contain only letters and spaces")
	validator.Field("age").Range(18, 120)
	validator.Field("email").Format(EmailPattern, "Invalid email format")
	
	// Test valid data
	validData := map[string]interface{}{
		"name":  "John Doe",
		"age":   25,
		"email": "john.doe@example.com",
	}
	
	err := validator.Validate(context.Background(), validData)
	if err != nil {
		t.Errorf("Expected no error for valid data, got: %v", err)
	}
	
	// Test invalid data
	invalidData := map[string]interface{}{
		"name":  "",           // Missing required
		"age":   15,           // Below range
		"email": "invalid",    // Invalid format
	}
	
	err = validator.Validate(context.Background(), invalidData)
	if err == nil {
		t.Error("Expected validation error for invalid data")
	}
}

func TestCommonPatterns(t *testing.T) {
	tests := []struct {
		pattern string
		valid   []string
		invalid []string
	}{
		{
			pattern: EmailPattern,
			valid:   []string{"test@example.com", "user.name@domain.co.uk"},
			invalid: []string{"invalid", "@domain.com", "user@"},
		},
		{
			pattern: URLPattern,
			valid:   []string{"https://example.com", "http://test.org/path"},
			invalid: []string{"not-a-url", "ftp://invalid"},
		},
		{
			pattern: APIKeyPattern,
			valid:   []string{"abcdef1234567890abcdef1234567890abcd", "1234567890abcdef1234567890abcdef1234"},
			invalid: []string{"short", "toolongapikeythatexceedslimit", "invalid-chars!"},
		},
	}
	
	for _, test := range tests {
		validator := NewFieldValidator("test_field")
		validator.Format(test.pattern, "Invalid format")
		
		for _, validValue := range test.valid {
			err := validator.Validate(validValue)
			if err != nil {
				t.Errorf("Expected valid value '%s' to pass for pattern %s, got: %v", validValue, test.pattern, err)
			}
		}
		
		for _, invalidValue := range test.invalid {
			err := validator.Validate(invalidValue)
			if err == nil {
				t.Errorf("Expected invalid value '%s' to fail for pattern %s", invalidValue, test.pattern)
			}
		}
	}
}

func TestValidationErrorMessage(t *testing.T) {
	validator := NewFieldValidator("test_field")
	validator.Required()
	
	err := validator.Validate("")
	if err == nil {
		t.Fatal("Expected validation error")
	}
	
	errMsg := err.Error()
	if errMsg == "" {
		t.Error("Expected non-empty error message")
	}
	
	// Check that the error message contains field name
	// Note: This depends on the actual implementation
}

func TestMultipleFieldValidation(t *testing.T) {
	validator := NewStructValidator()
	validator.Field("field1").Required()
	validator.Field("field2").Required()
	validator.Field("field3").Range(1, 100)
	
	// Test data with multiple validation errors
	invalidData := map[string]interface{}{
		"field1": "",    // Missing required
		"field2": nil,   // Missing required  
		"field3": 200,   // Out of range
	}
	
	err := validator.Validate(context.Background(), invalidData)
	if err == nil {
		t.Error("Expected validation error for multiple invalid fields")
	}
	
	// The error should ideally contain information about all failed validations
	errMsg := err.Error()
	if errMsg == "" {
		t.Error("Expected non-empty error message")
	}
}

func TestRangeWithFloats(t *testing.T) {
	validator := NewFieldValidator("test_field")
	validator.Range(1.5, 10.5)
	
	// Test float values
	err := validator.Validate(5.5)
	if err != nil {
		t.Errorf("Expected no error for valid float value, got: %v", err)
	}
	
	err = validator.Validate(0.5)
	if err == nil {
		t.Error("Expected validation error for float below range")
	}
	
	err = validator.Validate(11.5)
	if err == nil {
		t.Error("Expected validation error for float above range")
	}
} 