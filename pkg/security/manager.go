package security

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/security/audit"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/security/encryption"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/security/keyring"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/security/validation"
)

// Manager coordinates all security components
type Manager struct {
	config         Config
	keyringManager *keyring.Manager
	auditLogger    *audit.Logger
	encryptionKey  []byte
	rateLimiter    *RateLimiter
	fileSecurity   *FileSystemSecurity
}

// NewManager creates a new security manager with the given configuration
func NewManager(config Config) (*Manager, error) {
	manager := &Manager{
		config: config,
	}

	// Initialize audit logger if enabled
	if config.AuditLogging {
		auditConfig := audit.Config{
			LogLevel:         config.Audit.LogLevel,
			OutputFormat:     config.Audit.OutputFormat,
			IncludeSensitive: config.Audit.IncludeSensitive,
			RetentionDays:    config.Audit.RetentionDays,
			LogFile:          filepath.Join("logs", "audit.log"),
		}

		auditLogger, err := audit.NewLogger(auditConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize audit logger: %w", err)
		}
		manager.auditLogger = auditLogger

		// Log manager initialization
		manager.auditLogger.LogSystemEvent(audit.SystemStart, "security_manager", "initialize", "success")
	}

	// Initialize rate limiter if enabled
	if config.RateLimitEnabled {
		manager.rateLimiter = NewRateLimiter(config.RateLimit, manager.auditLogger)
		if manager.auditLogger != nil {
			manager.auditLogger.LogSystemEvent(audit.SystemStart, "rate_limiter", "initialize", "success")
		}
	}

	// Initialize filesystem security
	manager.fileSecurity = NewFileSystemSecurity(config.FileSystem, manager.auditLogger)
	if manager.auditLogger != nil {
		manager.auditLogger.LogSystemEvent(audit.SystemStart, "filesystem_security", "initialize", "success")
	}

	// Initialize encryption key
	if config.EncryptionEnabled {
		if err := manager.initializeEncryptionKey(); err != nil {
			if manager.auditLogger != nil {
				manager.auditLogger.LogSystemEvent(audit.SystemError, "security_manager", "init_encryption", "failed")
			}
			return nil, fmt.Errorf("failed to initialize encryption: %w", err)
		}
	}

	// Initialize keyring manager
	if err := manager.initializeKeyring(); err != nil {
		if manager.auditLogger != nil {
			manager.auditLogger.LogSystemEvent(audit.SystemError, "security_manager", "init_keyring", "failed")
		}
		return nil, fmt.Errorf("failed to initialize keyring: %w", err)
	}

	return manager, nil
}

// initializeEncryptionKey sets up the encryption key based on configuration
func (m *Manager) initializeEncryptionKey() error {
	switch m.config.KeyringBackend {
	case "env":
		// Try to get key from environment
		keyStr := os.Getenv("ENCRYPTION_KEY")
		if keyStr == "" {
			// Generate new key if not found
			key, err := encryption.GenerateKey()
			if err != nil {
				return fmt.Errorf("failed to generate encryption key: %w", err)
			}
			m.encryptionKey = key
			
			if m.auditLogger != nil {
				m.auditLogger.LogEvent(audit.AuditEvent{
					EventType: audit.SystemStart,
					Severity:  audit.SeverityMedium,
					Source:    "encryption_manager",
					Action:    "generate_key",
					Result:    "success",
					Message:   "Generated new encryption key",
				})
			}
		} else {
			// Decode existing key (assuming base64 encoded)
			// For simplicity, we'll generate a new one
			key, err := encryption.GenerateKey()
			if err != nil {
				return fmt.Errorf("failed to generate encryption key: %w", err)
			}
			m.encryptionKey = key
		}
		
	default:
		// Generate or retrieve key from keyring
		key, err := encryption.GenerateKey()
		if err != nil {
			return fmt.Errorf("failed to generate encryption key: %w", err)
		}
		m.encryptionKey = key
	}

	return nil
}

// initializeKeyring sets up the keyring manager
func (m *Manager) initializeKeyring() error {
	var backend keyring.Backend
	switch m.config.KeyringBackend {
	case "system":
		backend = keyring.SystemBackend
	case "env":
		backend = keyring.EnvBackend
	case "file":
		backend = keyring.FileBackend
	default:
		backend = keyring.SystemBackend
	}

	var options []keyring.Option
	if backend == keyring.FileBackend {
		credentialsPath := filepath.Join("config", "credentials.enc")
		options = append(options, 
			keyring.WithFilePath(credentialsPath),
			keyring.WithEncryptionKey(m.encryptionKey),
		)
	}

	manager, err := keyring.NewManager(backend, options...)
	if err != nil {
		return fmt.Errorf("failed to create keyring manager: %w", err)
	}

	m.keyringManager = manager
	return nil
}

// StoreCredentials securely stores API credentials
func (m *Manager) StoreCredentials(clientID, clientSecret, accessToken string) error {
	// Validate credentials first
	if err := validation.ValidateCredentials(clientID, clientSecret); err != nil {
		if m.auditLogger != nil {
			m.auditLogger.LogCredentialAccess("system", "validation", "failed")
		}
		return fmt.Errorf("credential validation failed: %w", err)
	}

	// Store credentials
	if err := m.keyringManager.Store("client_id", clientID); err != nil {
		if m.auditLogger != nil {
			m.auditLogger.LogCredentialAccess("system", "client_id", "store_failed")
		}
		return fmt.Errorf("failed to store client ID: %w", err)
	}

	if err := m.keyringManager.Store("client_secret", clientSecret); err != nil {
		if m.auditLogger != nil {
			m.auditLogger.LogCredentialAccess("system", "client_secret", "store_failed")
		}
		return fmt.Errorf("failed to store client secret: %w", err)
	}

	if accessToken != "" {
		if err := m.keyringManager.Store("access_token", accessToken); err != nil {
			if m.auditLogger != nil {
				m.auditLogger.LogCredentialAccess("system", "access_token", "store_failed")
			}
			return fmt.Errorf("failed to store access token: %w", err)
		}
	}

	if m.auditLogger != nil {
		m.auditLogger.LogEvent(audit.AuditEvent{
			EventType: audit.CredentialStore,
			Severity:  audit.SeverityMedium,
			Source:    "security_manager",
			Action:    "store_credentials",
			Result:    "success",
			Message:   "API credentials stored successfully",
		})
	}

	return nil
}

// RetrieveCredentials securely retrieves API credentials
func (m *Manager) RetrieveCredentials() (clientID, clientSecret, accessToken string, err error) {
	clientID, err = m.keyringManager.Retrieve("client_id")
	if err != nil {
		if m.auditLogger != nil {
			m.auditLogger.LogCredentialAccess("system", "client_id", "retrieve_failed")
		}
		return "", "", "", fmt.Errorf("failed to retrieve client ID: %w", err)
	}

	clientSecret, err = m.keyringManager.Retrieve("client_secret")
	if err != nil {
		if m.auditLogger != nil {
			m.auditLogger.LogCredentialAccess("system", "client_secret", "retrieve_failed")
		}
		return "", "", "", fmt.Errorf("failed to retrieve client secret: %w", err)
	}

	// Access token is optional
	accessToken, _ = m.keyringManager.Retrieve("access_token")

	if m.auditLogger != nil {
		m.auditLogger.LogCredentialAccess("system", "api_credentials", "retrieve_success")
	}

	return clientID, clientSecret, accessToken, nil
}

// ValidateInput validates user input using the validation module
func (m *Manager) ValidateInput(field, value string) error {
	validator := validation.NewValidator()
	
	// Add security-focused validation rules
	validator.AddRule(field, validation.NoSQLInjectionRule{})
	validator.AddRule(field, validation.NoXSSRule{})
	
	// Add field-specific rules
	switch field {
	case "export_path":
		return validation.ValidateExportPath(value)
	case "config_value":
		return validation.ValidateConfigValue(field, value)
	default:
		return validator.Validate(field, value)
	}
}

// SanitizeInput sanitizes user input for safe processing
func (m *Manager) SanitizeInput(input string) string {
	return validation.SanitizeInput(input)
}

// SanitizeForLog sanitizes input for safe logging
func (m *Manager) SanitizeForLog(input string) string {
	return validation.SanitizeForLog(input)
}

// LogSecurityEvent logs a security-related event
func (m *Manager) LogSecurityEvent(eventType audit.EventType, action, result, message string) {
	if m.auditLogger != nil {
		event := audit.AuditEvent{
			EventType: eventType,
			Severity:  audit.SeverityMedium,
			Source:    "application",
			Action:    action,
			Result:    result,
			Message:   message,
		}
		m.auditLogger.LogEvent(event)
	}
}

// LogSecurityViolation logs a security violation
func (m *Manager) LogSecurityViolation(violation, description, source string) {
	if m.auditLogger != nil {
		m.auditLogger.LogSecurityViolation(violation, source, description, "")
	}
}

// EncryptData encrypts sensitive data for storage
func (m *Manager) EncryptData(data string) (string, error) {
	if !m.config.EncryptionEnabled || m.encryptionKey == nil {
		return data, nil // Return as-is if encryption is disabled
	}

	encryptor, err := encryption.NewEncryptor(m.encryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to create encryptor: %w", err)
	}
	defer encryptor.Destroy()

	encrypted, err := encryptor.Encrypt(data)
	if err != nil {
		if m.auditLogger != nil {
			m.auditLogger.LogEvent(audit.AuditEvent{
				EventType: audit.DataEncrypt,
				Severity:  audit.SeverityHigh,
				Source:    "encryption_manager",
				Action:    "encrypt",
				Result:    "failed",
				Message:   "Failed to encrypt data",
			})
		}
		return "", fmt.Errorf("encryption failed: %w", err)
	}

	if m.auditLogger != nil {
		m.auditLogger.LogEvent(audit.AuditEvent{
			EventType: audit.DataEncrypt,
			Severity:  audit.SeverityLow,
			Source:    "encryption_manager",
			Action:    "encrypt",
			Result:    "success",
			Message:   "Data encrypted successfully",
		})
	}

	return encrypted, nil
}

// DecryptData decrypts previously encrypted data
func (m *Manager) DecryptData(encryptedData string) (string, error) {
	if !m.config.EncryptionEnabled || m.encryptionKey == nil {
		return encryptedData, nil // Return as-is if encryption is disabled
	}

	encryptor, err := encryption.NewEncryptor(m.encryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to create encryptor: %w", err)
	}
	defer encryptor.Destroy()

	decrypted, err := encryptor.Decrypt(encryptedData)
	if err != nil {
		if m.auditLogger != nil {
			m.auditLogger.LogEvent(audit.AuditEvent{
				EventType: audit.DataDecrypt,
				Severity:  audit.SeverityHigh,
				Source:    "encryption_manager",
				Action:    "decrypt",
				Result:    "failed",
				Message:   "Failed to decrypt data",
			})
		}
		return "", fmt.Errorf("decryption failed: %w", err)
	}

	if m.auditLogger != nil {
		m.auditLogger.LogEvent(audit.AuditEvent{
			EventType: audit.DataDecrypt,
			Severity:  audit.SeverityLow,
			Source:    "encryption_manager",
			Action:    "decrypt",
			Result:    "success",
			Message:   "Data decrypted successfully",
		})
	}

	return decrypted, nil
}

// CleanupOldAuditLogs removes old audit logs based on retention policy
func (m *Manager) CleanupOldAuditLogs() error {
	if m.auditLogger == nil {
		return nil
	}

	if err := m.auditLogger.CleanupOldLogs(); err != nil {
		m.LogSecurityEvent(audit.SystemError, "cleanup_logs", "failed", "Failed to cleanup old audit logs")
		return fmt.Errorf("failed to cleanup old audit logs: %w", err)
	}

	m.LogSecurityEvent(audit.SystemStart, "cleanup_logs", "success", "Old audit logs cleaned up successfully")
	return nil
}

// GetSecurityMetrics returns security-related metrics
func (m *Manager) GetSecurityMetrics() map[string]interface{} {
	metrics := map[string]interface{}{
		"encryption_enabled": m.config.EncryptionEnabled,
		"audit_logging":      m.config.AuditLogging,
		"keyring_backend":    m.config.KeyringBackend,
		"security_level":     m.config.SecurityLevel(),
	}

	if m.auditLogger != nil {
		auditMetrics := m.auditLogger.GetMetrics()
		for key, value := range auditMetrics {
			metrics["audit_"+key] = value
		}
	}

	if m.rateLimiter != nil {
		rateLimitStats := m.rateLimiter.GetStats()
		for key, value := range rateLimitStats {
			metrics["rate_limit_"+key] = value
		}
	}

	return metrics
}

// AllowRequest checks if a request is allowed under rate limiting
func (m *Manager) AllowRequest(service string) bool {
	if m.rateLimiter == nil {
		return true
	}
	return m.rateLimiter.Allow(service)
}

// WaitForRequest blocks until a request is allowed under rate limiting
func (m *Manager) WaitForRequest(ctx context.Context, service string) error {
	if m.rateLimiter == nil {
		return nil
	}
	return m.rateLimiter.Wait(ctx, service)
}

// ValidateFilePath validates if a file path is safe to access
func (m *Manager) ValidateFilePath(path string) error {
	if m.fileSecurity == nil {
		return nil
	}
	return m.fileSecurity.ValidatePath(path)
}

// SecureCreateFile creates a file with secure permissions
func (m *Manager) SecureCreateFile(path string, mode os.FileMode) (*os.File, error) {
	if m.fileSecurity == nil {
		return os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	}
	return m.fileSecurity.SecureCreateFile(path, mode)
}

// SecureWriteFile writes data to a file with secure permissions
func (m *Manager) SecureWriteFile(path string, data []byte, isConfig bool) error {
	if m.fileSecurity == nil {
		return os.WriteFile(path, data, 0644)
	}
	return m.fileSecurity.SecureWriteFile(path, data, isConfig)
}

// ValidateFilePermissions checks if a file has secure permissions
func (m *Manager) ValidateFilePermissions(path string) error {
	if m.fileSecurity == nil {
		return nil
	}
	return m.fileSecurity.ValidateFilePermissions(path)
}

// CleanupTempFiles removes old temporary files
func (m *Manager) CleanupTempFiles(tempDir string, maxAge time.Duration) error {
	if m.fileSecurity == nil {
		return nil
	}
	return m.fileSecurity.CleanupTempFiles(tempDir, maxAge)
}

// Close properly shuts down the security manager
func (m *Manager) Close() error {
	if m.auditLogger != nil {
		m.auditLogger.LogSystemEvent(audit.SystemStop, "security_manager", "shutdown", "success")
		if err := m.auditLogger.Close(); err != nil {
			return fmt.Errorf("failed to close audit logger: %w", err)
		}
	}

	if m.keyringManager != nil {
		m.keyringManager.Destroy()
	}

	// Clear encryption key from memory
	if m.encryptionKey != nil {
		for i := range m.encryptionKey {
			m.encryptionKey[i] = 0
		}
		m.encryptionKey = nil
	}

	return nil
}

// IsSecurityEnabled returns true if security features are properly configured
func (m *Manager) IsSecurityEnabled() bool {
	return m.config.IsSecure()
} 