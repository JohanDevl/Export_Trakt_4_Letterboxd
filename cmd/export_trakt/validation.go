package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/config"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/logger"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/security"
)

// validateSecurityConfiguration performs comprehensive security validation
func validateSecurityConfiguration(cfg *config.Config, log logger.Logger) int {
	fmt.Println("🔒 Security Configuration Validation")
	fmt.Println("=====================================")

	var errors []string
	var warnings []string

	// 1. Validate security configuration
	if err := cfg.Security.Validate(); err != nil {
		errors = append(errors, fmt.Sprintf("Security config validation failed: %v", err))
	} else {
		fmt.Println("✅ Security configuration is valid")
	}

	// 2. Check security level
	securityLevel := cfg.Security.SecurityLevel()
	switch securityLevel {
	case "high":
		fmt.Println("✅ Security level: HIGH - All security features enabled")
	case "medium":
		fmt.Println("⚠️  Security level: MEDIUM - Some security features disabled")
		warnings = append(warnings, "Consider enabling all security features for production use")
	case "low":
		fmt.Println("❌ Security level: LOW - Critical security features disabled")
		errors = append(errors, "Security level is too low for production use")
	}

	// 3. Test security manager initialization
	fmt.Println("\n🔧 Testing Security Manager...")
	securityManager, err := security.NewManager(cfg.Security)
	if err != nil {
		errors = append(errors, fmt.Sprintf("Security manager initialization failed: %v", err))
	} else {
		fmt.Println("✅ Security manager initialized successfully")

		// Test encryption if enabled
		if cfg.Security.EncryptionEnabled {
			testData := "test-encryption-data"
			encrypted, err := securityManager.EncryptData(testData)
			if err != nil {
				errors = append(errors, fmt.Sprintf("Encryption test failed: %v", err))
			} else {
				decrypted, err := securityManager.DecryptData(encrypted)
				if err != nil {
					errors = append(errors, fmt.Sprintf("Decryption test failed: %v", err))
				} else if decrypted != testData {
					errors = append(errors, "Encryption/decryption round-trip failed")
				} else {
					fmt.Println("✅ Encryption/decryption test passed")
				}
			}
		}

		// Test input validation
		testInput := "<script>alert('xss')</script>"
		sanitized := securityManager.SanitizeInput(testInput)
		if sanitized == testInput {
			warnings = append(warnings, "Input sanitization may not be working properly")
		} else {
			fmt.Println("✅ Input sanitization working")
		}

		// Test file path validation
		maliciousPath := "../../../etc/passwd"
		if err := securityManager.ValidateFilePath(maliciousPath); err == nil {
			errors = append(errors, "Path traversal protection not working")
		} else {
			fmt.Println("✅ Path traversal protection working")
		}

		// Clean up
		if err := securityManager.Close(); err != nil {
			warnings = append(warnings, fmt.Sprintf("Security manager cleanup warning: %v", err))
		}
	}

	// 4. Check file permissions
	fmt.Println("\n📁 Checking File Permissions...")
	configFile := "config/config.toml"
	if info, err := os.Stat(configFile); err == nil {
		mode := info.Mode()
		if mode&0077 != 0 {
			warnings = append(warnings, fmt.Sprintf("Config file %s has overly permissive permissions: %v", configFile, mode))
		} else {
			fmt.Println("✅ Config file permissions are secure")
		}
	} else {
		fmt.Printf("ℹ️  Config file %s not found (using defaults)\n", configFile)
	}

	// 5. Check credential storage
	fmt.Println("\n🔑 Checking Credential Storage...")
	switch cfg.Security.KeyringBackend {
	case "system":
		fmt.Println("✅ Using system keyring (most secure)")
	case "env":
		fmt.Println("⚠️  Using environment variables for credentials")
		warnings = append(warnings, "Environment variables are less secure than system keyring")

		// Check if credentials are in config file
		if cfg.Trakt.ClientID != "" || cfg.Trakt.ClientSecret != "" {
			errors = append(errors, "Credentials found in config file while using env backend")
		}
	case "file":
		fmt.Println("⚠️  Using encrypted file for credentials")
		warnings = append(warnings, "File-based credential storage is less secure than system keyring")
	default:
		errors = append(errors, fmt.Sprintf("Unknown keyring backend: %s", cfg.Security.KeyringBackend))
	}

	// 6. Check HTTPS enforcement
	fmt.Println("\n🌐 Checking HTTPS Configuration...")
	if cfg.Security.RequireHTTPS {
		fmt.Println("✅ HTTPS enforcement enabled")

		// Check if API URL uses HTTPS
		if !strings.HasPrefix(cfg.Trakt.APIBaseURL, "https://") {
			errors = append(errors, "API base URL must use HTTPS when HTTPS enforcement is enabled")
		}
	} else {
		warnings = append(warnings, "HTTPS enforcement is disabled")
	}

	// 7. Check audit logging
	fmt.Println("\n📝 Checking Audit Configuration...")
	if cfg.Security.AuditLogging {
		fmt.Println("✅ Audit logging enabled")

		if cfg.Security.Audit.IncludeSensitive {
			warnings = append(warnings, "Audit logging includes sensitive information (not recommended for production)")
		}

		if cfg.Security.Audit.RetentionDays < 30 {
			warnings = append(warnings, "Audit log retention period is less than 30 days")
		}
	} else {
		warnings = append(warnings, "Audit logging is disabled")
	}

	// 8. Check rate limiting
	fmt.Println("\n🚦 Checking Rate Limiting...")
	if cfg.Security.RateLimitEnabled {
		fmt.Println("✅ Rate limiting enabled")
	} else {
		warnings = append(warnings, "Rate limiting is disabled")
	}

	// 9. Display summary
	fmt.Println("\n📊 Security Validation Summary")
	fmt.Println("==============================")

	if len(errors) == 0 && len(warnings) == 0 {
		fmt.Println("🎉 All security checks passed!")
		log.Info("security.validation_success", nil)
		return 0
	}

	if len(warnings) > 0 {
		fmt.Printf("⚠️  %d Warning(s):\n", len(warnings))
		for i, warning := range warnings {
			fmt.Printf("   %d. %s\n", i+1, warning)
		}
		fmt.Println()
	}

	if len(errors) > 0 {
		fmt.Printf("❌ %d Error(s):\n", len(errors))
		for i, error := range errors {
			fmt.Printf("   %d. %s\n", i+1, error)
		}
		fmt.Println()

		log.Error("security.validation_failed", map[string]interface{}{
			"error_count":   len(errors),
			"warning_count": len(warnings),
		})

		fmt.Println("🔒 Security validation failed. Please fix the errors above.")
		return 1
	}

	log.Info("security.validation_warning", map[string]interface{}{
		"warning_count": len(warnings),
	})

	fmt.Println("⚠️  Security validation completed with warnings. Review recommendations above.")
	return 0
}
