package security

import (
	"fmt"
	"strings"
)

// Config holds all security-related configuration
type Config struct {
	EncryptionEnabled  bool         `toml:"encryption_enabled"`
	KeyringBackend     string       `toml:"keyring_backend"`     // system, env, file
	AuditLogging       bool         `toml:"audit_logging"`
	RateLimitEnabled   bool         `toml:"rate_limit_enabled"`
	RequireHTTPS       bool         `toml:"require_https"`
	Audit              AuditConfig  `toml:"audit"`
	RateLimit          RateLimitConfig `toml:"rate_limit"`
	FileSystem         FileSystemConfig `toml:"filesystem"`
	HTTPS              HTTPSConfig  `toml:"https"`
}

// AuditConfig holds audit logging configuration
type AuditConfig struct {
	LogLevel        string `toml:"log_level"`        // debug, info, warn, error
	RetentionDays   int    `toml:"retention_days"`   // 90 days default
	IncludeSensitive bool   `toml:"include_sensitive"` // false default for security
	OutputFormat    string `toml:"output_format"`    // json, text
}

// DefaultSecurityConfig returns a secure default configuration
func DefaultSecurityConfig() Config {
	return Config{
		EncryptionEnabled:  true,
		KeyringBackend:     "system",
		AuditLogging:       true,
		RateLimitEnabled:   true,
		RequireHTTPS:       true,
		Audit: AuditConfig{
			LogLevel:         "info",
			RetentionDays:    90,
			IncludeSensitive: false,
			OutputFormat:     "json",
		},
		RateLimit:  DefaultRateLimitConfig(),
		FileSystem: DefaultFileSystemConfig(),
		HTTPS:      DefaultHTTPSConfig(),
	}
}

// Validate checks if the security configuration is valid
func (c *Config) Validate() error {
	if err := c.validateKeyringBackend(); err != nil {
		return fmt.Errorf("keyring backend: %w", err)
	}

	if err := c.Audit.Validate(); err != nil {
		return fmt.Errorf("audit config: %w", err)
	}

	return nil
}

// validateKeyringBackend validates the keyring backend setting
func (c *Config) validateKeyringBackend() error {
	validBackends := []string{"system", "env", "file"}
	backend := strings.ToLower(c.KeyringBackend)
	
	for _, valid := range validBackends {
		if backend == valid {
			return nil
		}
	}
	
	return fmt.Errorf("invalid keyring backend: %s (must be one of: %s)", 
		c.KeyringBackend, strings.Join(validBackends, ", "))
}

// Validate checks if the audit configuration is valid
func (ac *AuditConfig) Validate() error {
	if err := ac.validateLogLevel(); err != nil {
		return fmt.Errorf("log level: %w", err)
	}

	if err := ac.validateRetentionDays(); err != nil {
		return fmt.Errorf("retention days: %w", err)
	}

	if err := ac.validateOutputFormat(); err != nil {
		return fmt.Errorf("output format: %w", err)
	}

	return nil
}

// validateLogLevel validates the audit log level setting
func (ac *AuditConfig) validateLogLevel() error {
	validLevels := []string{"debug", "info", "warn", "error"}
	level := strings.ToLower(ac.LogLevel)
	
	for _, valid := range validLevels {
		if level == valid {
			return nil
		}
	}
	
	return fmt.Errorf("invalid log level: %s (must be one of: %s)", 
		ac.LogLevel, strings.Join(validLevels, ", "))
}

// validateRetentionDays validates the retention days setting
func (ac *AuditConfig) validateRetentionDays() error {
	if ac.RetentionDays < 1 {
		return fmt.Errorf("retention days must be positive, got: %d", ac.RetentionDays)
	}
	
	if ac.RetentionDays > 3650 { // 10 years max
		return fmt.Errorf("retention days too high (max 3650): %d", ac.RetentionDays)
	}
	
	return nil
}

// validateOutputFormat validates the output format setting
func (ac *AuditConfig) validateOutputFormat() error {
	validFormats := []string{"json", "text"}
	format := strings.ToLower(ac.OutputFormat)
	
	for _, valid := range validFormats {
		if format == valid {
			return nil
		}
	}
	
	return fmt.Errorf("invalid output format: %s (must be one of: %s)", 
		ac.OutputFormat, strings.Join(validFormats, ", "))
}

// IsSecure returns true if the configuration meets minimum security requirements
func (c *Config) IsSecure() bool {
	return c.EncryptionEnabled && 
		   c.AuditLogging && 
		   c.RequireHTTPS && 
		   c.KeyringBackend != "file" // file backend is less secure
}

// SecurityLevel returns a string describing the current security level
func (c *Config) SecurityLevel() string {
	if c.IsSecure() {
		return "high"
	}
	
	if c.EncryptionEnabled && c.RequireHTTPS {
		return "medium"
	}
	
	return "low"
} 