package validation

import (
	"fmt"
	"html"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Value   string
	Rule    string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation failed for field '%s': %s", e.Field, e.Message)
}

// Validator provides input validation capabilities
type Validator struct {
	rules map[string][]Rule
}

// Rule represents a validation rule
type Rule interface {
	Validate(value string) error
	Name() string
}

// NewValidator creates a new input validator
func NewValidator() *Validator {
	return &Validator{
		rules: make(map[string][]Rule),
	}
}

// AddRule adds a validation rule for a specific field
func (v *Validator) AddRule(field string, rule Rule) {
	v.rules[field] = append(v.rules[field], rule)
}

// Validate validates a field value against its configured rules
func (v *Validator) Validate(field, value string) error {
	rules, exists := v.rules[field]
	if !exists {
		return nil // No rules configured for this field
	}

	for _, rule := range rules {
		if err := rule.Validate(value); err != nil {
			return &ValidationError{
				Field:   field,
				Value:   value,
				Rule:    rule.Name(),
				Message: err.Error(),
			}
		}
	}

	return nil
}

// ValidateStruct validates multiple fields at once
func (v *Validator) ValidateStruct(data map[string]string) []error {
	var errors []error

	for field, value := range data {
		if err := v.Validate(field, value); err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

// Pre-defined validation rules

// RequiredRule ensures the field is not empty
type RequiredRule struct{}

func (r RequiredRule) Validate(value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("field is required")
	}
	return nil
}

func (r RequiredRule) Name() string {
	return "required"
}

// LengthRule validates string length
type LengthRule struct {
	Min int
	Max int
}

func (r LengthRule) Validate(value string) error {
	length := len(value)
	if length < r.Min {
		return fmt.Errorf("minimum length is %d characters", r.Min)
	}
	if r.Max > 0 && length > r.Max {
		return fmt.Errorf("maximum length is %d characters", r.Max)
	}
	return nil
}

func (r LengthRule) Name() string {
	return "length"
}

// RegexRule validates against a regular expression
type RegexRule struct {
	Pattern *regexp.Regexp
	Message string
}

func (r RegexRule) Validate(value string) error {
	if !r.Pattern.MatchString(value) {
		if r.Message != "" {
			return fmt.Errorf(r.Message)
		}
		return fmt.Errorf("value does not match required pattern")
	}
	return nil
}

func (r RegexRule) Name() string {
	return "regex"
}

// AlphanumericRule ensures only alphanumeric characters
type AlphanumericRule struct {
	AllowSpaces bool
	AllowHyphens bool
	AllowUnderscores bool
}

func (r AlphanumericRule) Validate(value string) error {
	for _, char := range value {
		if unicode.IsLetter(char) || unicode.IsDigit(char) {
			continue
		}
		if r.AllowSpaces && unicode.IsSpace(char) {
			continue
		}
		if r.AllowHyphens && char == '-' {
			continue
		}
		if r.AllowUnderscores && char == '_' {
			continue
		}
		return fmt.Errorf("contains invalid character: %c", char)
	}
	return nil
}

func (r AlphanumericRule) Name() string {
	return "alphanumeric"
}

// EmailRule validates email format
type EmailRule struct{}

func (r EmailRule) Validate(value string) error {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(value) {
		return fmt.Errorf("invalid email format")
	}
	return nil
}

func (r EmailRule) Name() string {
	return "email"
}

// URLRule validates URL format
type URLRule struct {
	RequireHTTPS bool
}

func (r URLRule) Validate(value string) error {
	parsedURL, err := url.Parse(value)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	if parsedURL.Scheme == "" {
		return fmt.Errorf("URL must include scheme (http/https)")
	}

	if r.RequireHTTPS && parsedURL.Scheme != "https" {
		return fmt.Errorf("URL must use HTTPS")
	}

	if parsedURL.Host == "" {
		return fmt.Errorf("URL must include host")
	}

	return nil
}

func (r URLRule) Name() string {
	return "url"
}

// PathRule validates file paths for security
type PathRule struct {
	AllowAbsolute bool
	AllowParentDir bool
	RestrictToDir string
}

func (r PathRule) Validate(value string) error {
	// Check for path traversal attempts
	if !r.AllowParentDir && strings.Contains(value, "..") {
		return fmt.Errorf("path traversal not allowed")
	}

	// Check for absolute paths
	if !r.AllowAbsolute && filepath.IsAbs(value) {
		return fmt.Errorf("absolute paths not allowed")
	}

	// Restrict to specific directory
	if r.RestrictToDir != "" {
		cleanPath := filepath.Clean(value)
		cleanRestrictDir := filepath.Clean(r.RestrictToDir)
		if !strings.HasPrefix(cleanPath, cleanRestrictDir) {
			return fmt.Errorf("path must be within %s", r.RestrictToDir)
		}
	}

	// Check for dangerous characters
	dangerousChars := []string{"|", "&", ";", "$", "`", "\\", "\"", "'", "<", ">"}
	for _, char := range dangerousChars {
		if strings.Contains(value, char) {
			return fmt.Errorf("path contains dangerous character: %s", char)
		}
	}

	return nil
}

func (r PathRule) Name() string {
	return "path"
}

// NoSQLInjectionRule prevents common SQL injection patterns
type NoSQLInjectionRule struct{}

func (r NoSQLInjectionRule) Validate(value string) error {
	// Common SQL injection patterns
	sqlPatterns := []string{
		"'", "\"", "--", "/*", "*/", "xp_", "sp_",
		"union", "select", "insert", "update", "delete",
		"drop", "create", "alter", "exec", "execute",
	}

	// NoSQL injection patterns
	nosqlPatterns := []string{
		"$ne", "$gt", "$lt", "$gte", "$lte", "$in", "$nin",
		"$regex", "$where", "$exists", "$not", "$or", "$and",
		"$nor", "$size", "$all", "$elemMatch", "$mod",
	}

	lowerValue := strings.ToLower(value)
	
	// Check SQL patterns
	for _, pattern := range sqlPatterns {
		if strings.Contains(lowerValue, pattern) {
			return fmt.Errorf("contains potentially dangerous pattern: %s", pattern)
		}
	}
	
	// Check NoSQL patterns
	for _, pattern := range nosqlPatterns {
		if strings.Contains(lowerValue, pattern) {
			return fmt.Errorf("contains potentially dangerous NoSQL pattern: %s", pattern)
		}
	}

	return nil
}

func (r NoSQLInjectionRule) Name() string {
	return "no_sql_injection"
}

// NoXSSRule prevents XSS injection patterns
type NoXSSRule struct{}

func (r NoXSSRule) Validate(value string) error {
	// Common XSS patterns
	xssPatterns := []string{
		"<script", "</script>", "javascript:", "vbscript:",
		"onload=", "onerror=", "onclick=", "onmouseover=",
		"<iframe", "<object", "<embed", "<link",
	}

	lowerValue := strings.ToLower(value)
	for _, pattern := range xssPatterns {
		if strings.Contains(lowerValue, pattern) {
			return fmt.Errorf("contains potentially dangerous XSS pattern: %s", pattern)
		}
	}

	return nil
}

func (r NoXSSRule) Name() string {
	return "no_xss"
}

// WhitelistRule ensures value is in allowed list
type WhitelistRule struct {
	AllowedValues []string
	CaseSensitive bool
}

func (r WhitelistRule) Validate(value string) error {
	compareValue := value
	if !r.CaseSensitive {
		compareValue = strings.ToLower(value)
	}

	for _, allowed := range r.AllowedValues {
		compareAllowed := allowed
		if !r.CaseSensitive {
			compareAllowed = strings.ToLower(allowed)
		}
		if compareValue == compareAllowed {
			return nil
		}
	}

	return fmt.Errorf("value not in allowed list: %s", value)
}

func (r WhitelistRule) Name() string {
	return "whitelist"
}

// Sanitizers for safe data handling

// SanitizeInput safely sanitizes user input
func SanitizeInput(input string) string {
	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")
	
	// Remove control characters except newline and tab
	var result strings.Builder
	for _, char := range input {
		if unicode.IsPrint(char) || char == '\n' || char == '\t' {
			result.WriteRune(char)
		}
	}
	input = result.String()
	
	// Use HTML escaping for safe output
	input = html.EscapeString(input)
	
	return input
}

// sanitizeXSS removes or escapes XSS patterns
func sanitizeXSS(input string) string {
	// Remove script tags and their content
	scriptRegex := regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`)
	input = scriptRegex.ReplaceAllString(input, "")
	
	// Remove dangerous HTML tags
	dangerousTags := []string{
		"script", "iframe", "object", "embed", "link", "meta",
		"style", "form", "input", "button", "textarea",
	}
	
	for _, tag := range dangerousTags {
		// Remove opening tags
		openTagRegex := regexp.MustCompile(`(?i)<` + tag + `[^>]*>`)
		input = openTagRegex.ReplaceAllString(input, "")
		
		// Remove closing tags
		closeTagRegex := regexp.MustCompile(`(?i)</` + tag + `>`)
		input = closeTagRegex.ReplaceAllString(input, "")
	}
	
	// Remove javascript: and vbscript: protocols
	protocolRegex := regexp.MustCompile(`(?i)(javascript|vbscript|data):\s*`)
	input = protocolRegex.ReplaceAllString(input, "")
	
	// Remove event handlers
	eventHandlers := []string{
		"onload", "onerror", "onclick", "onmouseover", "onmouseout",
		"onkeydown", "onkeyup", "onkeypress", "onfocus", "onblur",
		"onsubmit", "onchange", "onselect", "onreset", "onabort",
	}
	
	for _, handler := range eventHandlers {
		handlerRegex := regexp.MustCompile(`(?i)` + handler + `\s*=\s*[^>\s]*`)
		input = handlerRegex.ReplaceAllString(input, "")
	}
	
	return input
}

// sanitizeSQL removes common SQL injection patterns
func sanitizeSQL(input string) string {
	// Remove SQL comment patterns
	input = strings.ReplaceAll(input, "--", "")
	input = strings.ReplaceAll(input, "/*", "")
	input = strings.ReplaceAll(input, "*/", "")
	
	// Remove dangerous SQL keywords (case-insensitive)
	dangerousKeywords := []string{
		"union", "select", "insert", "update", "delete",
		"drop", "create", "alter", "exec", "execute",
		"xp_", "sp_",
	}
	
	for _, keyword := range dangerousKeywords {
		// Use word boundaries to avoid breaking legitimate words
		keywordRegex := regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(keyword) + `\b`)
		input = keywordRegex.ReplaceAllString(input, "")
	}
	
	return input
}

// SanitizeForLog sanitizes input for safe logging
func SanitizeForLog(input string) string {
	// Replace newlines and carriage returns with spaces  
	input = strings.ReplaceAll(input, "\n", " ")
	input = strings.ReplaceAll(input, "\r", " ")
	input = strings.ReplaceAll(input, "\t", " ")
	
	// Replace control characters with spaces
	var result strings.Builder
	for _, char := range input {
		if unicode.IsPrint(char) || unicode.IsSpace(char) {
			result.WriteRune(char)
		} else {
			result.WriteRune(' ')
		}
	}
	
	output := result.String()
	
	// Normalize multiple spaces to single space
	spaceRegex := regexp.MustCompile(`\s+`)
	output = spaceRegex.ReplaceAllString(output, " ")
	
	return output
}

// SanitizeFilename sanitizes filename for safe file operations
func SanitizeFilename(filename string) string {
	// Handle specific pattern for directory traversal
	if strings.Contains(filename, "../") {
		// Replace each "../" sequence with "__" 
		filename = strings.ReplaceAll(filename, "../", "__")
		// Then replace any remaining dangerous characters
		dangerousChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|", " ", "."}
		for _, char := range dangerousChars {
			filename = strings.ReplaceAll(filename, char, "_")
		}
	} else {
		// Check for really dangerous characters that require full sanitization
		reallyDangerousChars := strings.ContainsAny(filename, "<>:\"|?*\\")
		
		if reallyDangerousChars {
			// Try to preserve extension by finding the last dot and what comes after
			lastDotIndex := strings.LastIndex(filename, ".")
			var base, ext string
			
			if lastDotIndex > 0 && lastDotIndex < len(filename)-1 {
				base = filename[:lastDotIndex]
				ext = filename[lastDotIndex+1:] // extension without the dot
			} else {
				base = filename
				ext = ""
			}
			
			// Replace * with double underscores first
			base = strings.ReplaceAll(base, "*", "__")
			
			// Replace other dangerous characters with single underscores
			otherDangerousChars := []string{"/", "\\", ":", "?", "\"", "<", ">", "|", " ", "."}
			for _, char := range otherDangerousChars {
				base = strings.ReplaceAll(base, char, "_")
			}
			
			// Clean the extension too
			if ext != "" {
				extDangerousChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|", " "}
				for _, char := range extDangerousChars {
					ext = strings.ReplaceAll(ext, char, "_")
				}
				// Add extra underscore before extension when there are really dangerous chars
				filename = base + "_." + ext
			} else {
				filename = base
			}
		} else {
			// For normal filenames, preserve extension dots but replace spaces and other chars
			dangerousChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|", " "}
			for _, char := range dangerousChars {
				filename = strings.ReplaceAll(filename, char, "_")
			}
			
			// Handle leading dots
			if strings.HasPrefix(filename, ".") {
				filename = "_" + filename[1:]
			}
		}
	}
	
	// Ensure not empty
	if filename == "" {
		filename = "file"
	}
	
	return filename
}

// ValidateCredentials validates API credentials
func ValidateCredentials(clientID, clientSecret string) error {
	validator := NewValidator()
	
	// Add rules for client ID
	validator.AddRule("client_id", RequiredRule{})
	validator.AddRule("client_id", LengthRule{Min: 3, Max: 256})
	validator.AddRule("client_id", AlphanumericRule{AllowHyphens: true, AllowUnderscores: true})
	
	// Add rules for client secret
	validator.AddRule("client_secret", RequiredRule{})
	validator.AddRule("client_secret", LengthRule{Min: 3, Max: 512})
	
	// Validate
	data := map[string]string{
		"client_id":     clientID,
		"client_secret": clientSecret,
	}
	
	errors := validator.ValidateStruct(data)
	if len(errors) > 0 {
		return errors[0] // Return first error
	}
	
	return nil
}

// ValidateExportPath validates export directory path
func ValidateExportPath(path string) error {
	validator := NewValidator()
	
	validator.AddRule("export_path", RequiredRule{})
	validator.AddRule("export_path", PathRule{
		AllowAbsolute:  false,
		AllowParentDir: false,
	})
	
	return validator.Validate("export_path", path)
}

// ValidateConfigValue validates configuration values based on type
func ValidateConfigValue(key, value string) error {
	validator := NewValidator()
	
	switch key {
	case "trakt.api_base_url":
		validator.AddRule(key, RequiredRule{})
		validator.AddRule(key, URLRule{RequireHTTPS: true})
	case "log_level":
		validator.AddRule(key, RequiredRule{})
		validator.AddRule(key, WhitelistRule{
			AllowedValues: []string{"debug", "info", "warn", "error"},
			CaseSensitive: false,
		})
	case "output_format":
		validator.AddRule(key, RequiredRule{})
		validator.AddRule(key, WhitelistRule{
			AllowedValues: []string{"csv", "json"},
			CaseSensitive: false,
		})
	case "i18n.language":
		validator.AddRule(key, RequiredRule{})
		validator.AddRule(key, RegexRule{
			Pattern: regexp.MustCompile(`^[a-z]{2}(-[A-Z]{2})?$`),
			Message: "language must be in format 'en' or 'en-US'",
		})
	default:
		// Generic validation for unknown keys
		validator.AddRule(key, NoSQLInjectionRule{})
		validator.AddRule(key, NoXSSRule{})
	}
	
	return validator.Validate(key, value)
} 