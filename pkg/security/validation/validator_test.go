package validation

import (
	"regexp"
	"testing"
)

func TestNewValidator(t *testing.T) {
	validator := NewValidator()
	if validator == nil {
		t.Fatal("NewValidator returned nil")
	}

	if validator.rules == nil {
		t.Fatal("Expected rules map to be initialized")
	}
}

func TestRequiredRule(t *testing.T) {
	rule := RequiredRule{}

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:    "valid non-empty value",
			value:   "test",
			wantErr: false,
		},
		{
			name:    "empty string",
			value:   "",
			wantErr: true,
		},
		{
			name:    "whitespace only",
			value:   "   ",
			wantErr: true,
		},
		{
			name:    "tab and newline",
			value:   "\t\n",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := rule.Validate(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("RequiredRule.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	if rule.Name() != "required" {
		t.Errorf("Expected rule name 'required', got %s", rule.Name())
	}
}

func TestLengthRule(t *testing.T) {
	rule := LengthRule{Min: 3, Max: 10}

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:    "valid length",
			value:   "test",
			wantErr: false,
		},
		{
			name:    "minimum length",
			value:   "abc",
			wantErr: false,
		},
		{
			name:    "maximum length",
			value:   "1234567890",
			wantErr: false,
		},
		{
			name:    "too short",
			value:   "ab",
			wantErr: true,
		},
		{
			name:    "too long",
			value:   "12345678901",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := rule.Validate(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("LengthRule.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	if rule.Name() != "length" {
		t.Errorf("Expected rule name 'length', got %s", rule.Name())
	}
}

func TestLengthRuleNoMax(t *testing.T) {
	rule := LengthRule{Min: 3, Max: 0} // No maximum

	err := rule.Validate("this is a very long string that should pass")
	if err != nil {
		t.Errorf("Expected no error for long string when max is 0, got: %v", err)
	}
}

func TestRegexRule(t *testing.T) {
	pattern := regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	rule := RegexRule{Pattern: pattern, Message: "Only alphanumeric characters allowed"}

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:    "valid alphanumeric",
			value:   "test123",
			wantErr: false,
		},
		{
			name:    "invalid with special chars",
			value:   "test@123",
			wantErr: true,
		},
		{
			name:    "invalid with spaces",
			value:   "test 123",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := rule.Validate(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("RegexRule.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	if rule.Name() != "regex" {
		t.Errorf("Expected rule name 'regex', got %s", rule.Name())
	}
}

func TestAlphanumericRule(t *testing.T) {
	rule := AlphanumericRule{
		AllowSpaces:      true,
		AllowHyphens:     true,
		AllowUnderscores: false,
	}

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:    "valid alphanumeric",
			value:   "test123",
			wantErr: false,
		},
		{
			name:    "valid with spaces",
			value:   "test 123",
			wantErr: false,
		},
		{
			name:    "valid with hyphens",
			value:   "test-123",
			wantErr: false,
		},
		{
			name:    "invalid with underscores",
			value:   "test_123",
			wantErr: true,
		},
		{
			name:    "invalid with special chars",
			value:   "test@123",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := rule.Validate(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("AlphanumericRule.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	if rule.Name() != "alphanumeric" {
		t.Errorf("Expected rule name 'alphanumeric', got %s", rule.Name())
	}
}

func TestEmailRule(t *testing.T) {
	rule := EmailRule{}

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:    "valid email",
			value:   "test@example.com",
			wantErr: false,
		},
		{
			name:    "valid email with subdomain",
			value:   "user@mail.example.com",
			wantErr: false,
		},
		{
			name:    "valid email with numbers",
			value:   "user123@example123.com",
			wantErr: false,
		},
		{
			name:    "invalid without @",
			value:   "testexample.com",
			wantErr: true,
		},
		{
			name:    "invalid without domain",
			value:   "test@",
			wantErr: true,
		},
		{
			name:    "invalid without TLD",
			value:   "test@example",
			wantErr: true,
		},
		{
			name:    "invalid with spaces",
			value:   "test @example.com",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := rule.Validate(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("EmailRule.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	if rule.Name() != "email" {
		t.Errorf("Expected rule name 'email', got %s", rule.Name())
	}
}

func TestURLRule(t *testing.T) {
	rule := URLRule{RequireHTTPS: true}

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:    "valid HTTPS URL",
			value:   "https://example.com",
			wantErr: false,
		},
		{
			name:    "valid HTTPS URL with path",
			value:   "https://api.example.com/v1/users",
			wantErr: false,
		},
		{
			name:    "invalid HTTP when HTTPS required",
			value:   "http://example.com",
			wantErr: true,
		},
		{
			name:    "invalid without scheme",
			value:   "example.com",
			wantErr: true,
		},
		{
			name:    "invalid malformed URL",
			value:   "not-a-url",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := rule.Validate(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("URLRule.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	if rule.Name() != "url" {
		t.Errorf("Expected rule name 'url', got %s", rule.Name())
	}
}

func TestURLRuleAllowHTTP(t *testing.T) {
	rule := URLRule{RequireHTTPS: false}

	err := rule.Validate("http://example.com")
	if err != nil {
		t.Errorf("Expected HTTP URL to be valid when HTTPS not required, got: %v", err)
	}
}

func TestPathRule(t *testing.T) {
	rule := PathRule{
		AllowAbsolute:  false,
		AllowParentDir: false,
		RestrictToDir:  "./config",
	}

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:    "valid relative path in allowed dir",
			value:   "./config/app.toml",
			wantErr: false,
		},
		{
			name:    "invalid absolute path",
			value:   "/etc/passwd",
			wantErr: true,
		},
		{
			name:    "invalid parent directory",
			value:   "../config/app.toml",
			wantErr: true,
		},
		{
			name:    "invalid outside restricted dir",
			value:   "./logs/app.log",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := rule.Validate(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("PathRule.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	if rule.Name() != "path" {
		t.Errorf("Expected rule name 'path', got %s", rule.Name())
	}
}

func TestNoSQLInjectionRule(t *testing.T) {
	rule := NoSQLInjectionRule{}

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:    "safe input",
			value:   "normal text",
			wantErr: false,
		},
		{
			name:    "SQL injection attempt",
			value:   "'; DROP TABLE users; --",
			wantErr: true,
		},
		{
			name:    "NoSQL injection attempt",
			value:   "{ $ne: null }",
			wantErr: true,
		},
		{
			name:    "script tag",
			value:   "<script>alert('xss')</script>",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := rule.Validate(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("NoSQLInjectionRule.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	if rule.Name() != "no_sql_injection" {
		t.Errorf("Expected rule name 'no_sql_injection', got %s", rule.Name())
	}
}

func TestNoXSSRule(t *testing.T) {
	rule := NoXSSRule{}

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:    "safe input",
			value:   "normal text",
			wantErr: false,
		},
		{
			name:    "script tag",
			value:   "<script>alert('xss')</script>",
			wantErr: true,
		},
		{
			name:    "javascript protocol",
			value:   "javascript:alert('xss')",
			wantErr: true,
		},
		{
			name:    "on event handler",
			value:   "onload=alert('xss')",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := rule.Validate(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("NoXSSRule.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	if rule.Name() != "no_xss" {
		t.Errorf("Expected rule name 'no_xss', got %s", rule.Name())
	}
}

func TestWhitelistRule(t *testing.T) {
	rule := WhitelistRule{
		AllowedValues: []string{"admin", "user", "guest"},
		CaseSensitive: true,
	}

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:    "valid value",
			value:   "admin",
			wantErr: false,
		},
		{
			name:    "invalid value",
			value:   "superuser",
			wantErr: true,
		},
		{
			name:    "case sensitive mismatch",
			value:   "Admin",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := rule.Validate(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("WhitelistRule.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	if rule.Name() != "whitelist" {
		t.Errorf("Expected rule name 'whitelist', got %s", rule.Name())
	}
}

func TestWhitelistRuleCaseInsensitive(t *testing.T) {
	rule := WhitelistRule{
		AllowedValues: []string{"admin", "user", "guest"},
		CaseSensitive: false,
	}

	err := rule.Validate("Admin")
	if err != nil {
		t.Errorf("Expected case insensitive match to pass, got: %v", err)
	}
}

func TestValidator(t *testing.T) {
	validator := NewValidator()

	// Add rules
	validator.AddRule("username", RequiredRule{})
	validator.AddRule("username", LengthRule{Min: 3, Max: 20})
	validator.AddRule("email", RequiredRule{})
	validator.AddRule("email", EmailRule{})

	tests := []struct {
		name    string
		field   string
		value   string
		wantErr bool
	}{
		{
			name:    "valid username",
			field:   "username",
			value:   "testuser",
			wantErr: false,
		},
		{
			name:    "invalid empty username",
			field:   "username",
			value:   "",
			wantErr: true,
		},
		{
			name:    "invalid short username",
			field:   "username",
			value:   "ab",
			wantErr: true,
		},
		{
			name:    "valid email",
			field:   "email",
			value:   "test@example.com",
			wantErr: false,
		},
		{
			name:    "invalid email",
			field:   "email",
			value:   "invalid-email",
			wantErr: true,
		},
		{
			name:    "field without rules",
			field:   "description",
			value:   "any value",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.field, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validator.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateStruct(t *testing.T) {
	validator := NewValidator()
	validator.AddRule("username", RequiredRule{})
	validator.AddRule("email", EmailRule{})

	data := map[string]string{
		"username": "testuser",
		"email":    "test@example.com",
		"bio":      "optional field",
	}

	errors := validator.ValidateStruct(data)
	if len(errors) != 0 {
		t.Errorf("Expected no errors for valid data, got: %v", errors)
	}

	// Test with invalid data
	invalidData := map[string]string{
		"username": "",
		"email":    "invalid-email",
	}

	errors = validator.ValidateStruct(invalidData)
	if len(errors) != 2 {
		t.Errorf("Expected 2 errors for invalid data, got: %d", len(errors))
	}
}

func TestSanitizeInput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal text",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "script tag",
			input:    "<script>alert('xss')</script>",
			expected: "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;",
		},
		{
			name:     "SQL injection",
			input:    "'; DROP TABLE users; --",
			expected: "&#39;; DROP TABLE users; --",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeInput(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeInput() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSanitizeForLog(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal text",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "with newlines",
			input:    "Hello\nWorld\r\n",
			expected: "Hello World ",
		},
		{
			name:     "with control characters",
			input:    "Hello\x00\x01World",
			expected: "Hello World",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeForLog(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeForLog() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal filename",
			input:    "document.txt",
			expected: "document.txt",
		},
		{
			name:     "filename with spaces",
			input:    "my document.txt",
			expected: "my_document.txt",
		},
		{
			name:     "filename with special chars",
			input:    "file<>:\"|?*.txt",
			expected: "file_________.txt",
		},
		{
			name:     "filename with path separators",
			input:    "../../../etc/passwd",
			expected: "______etc_passwd",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeFilename(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeFilename() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestValidateCredentials(t *testing.T) {
	tests := []struct {
		name         string
		clientID     string
		clientSecret string
		wantErr      bool
	}{
		{
			name:         "valid credentials",
			clientID:     "valid_client_id_123",
			clientSecret: "valid_client_secret_456",
			wantErr:      false,
		},
		{
			name:         "empty client ID",
			clientID:     "",
			clientSecret: "valid_client_secret_456",
			wantErr:      true,
		},
		{
			name:         "empty client secret",
			clientID:     "valid_client_id_123",
			clientSecret: "",
			wantErr:      true,
		},
		{
			name:         "short client ID",
			clientID:     "ab",
			clientSecret: "valid_client_secret_456",
			wantErr:      true,
		},
		{
			name:         "short client secret",
			clientID:     "valid_client_id_123",
			clientSecret: "ab",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCredentials(tt.clientID, tt.clientSecret)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCredentials() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateExportPath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "valid export path",
			path:    "./exports/data.csv",
			wantErr: false,
		},
		{
			name:    "invalid absolute path",
			path:    "/etc/passwd",
			wantErr: true,
		},
		{
			name:    "invalid parent directory",
			path:    "../../../etc/passwd",
			wantErr: true,
		},
		{
			name:    "empty path",
			path:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateExportPath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateExportPath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateConfigValue(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		value   string
		wantErr bool
	}{
		{
			name:    "valid log level",
			key:     "log_level",
			value:   "info",
			wantErr: false,
		},
		{
			name:    "invalid log level",
			key:     "log_level",
			value:   "invalid",
			wantErr: true,
		},
		{
			name:    "valid output format",
			key:     "output_format",
			value:   "json",
			wantErr: false,
		},
		{
			name:    "invalid output format",
			key:     "output_format",
			value:   "xml",
			wantErr: true,
		},
		{
			name:    "unknown key",
			key:     "unknown_key",
			value:   "any_value",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfigValue(tt.key, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfigValue() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidationError(t *testing.T) {
	err := ValidationError{
		Field:   "username",
		Value:   "ab",
		Rule:    "length",
		Message: "minimum length is 3 characters",
	}

	expected := "validation failed for field 'username': minimum length is 3 characters"
	if err.Error() != expected {
		t.Errorf("ValidationError.Error() = %v, want %v", err.Error(), expected)
	}
} 