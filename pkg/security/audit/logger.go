package audit

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// EventType represents the type of security event
type EventType string

const (
	// Authentication events
	AuthSuccess    EventType = "auth_success"
	AuthFailure    EventType = "auth_failure"
	AuthLogout     EventType = "auth_logout"
	
	// Credential events
	CredentialAccess   EventType = "credential_access"
	CredentialStore    EventType = "credential_store"
	CredentialDelete   EventType = "credential_delete"
	CredentialRotation EventType = "credential_rotation"
	
	// Data events
	DataExport    EventType = "data_export"
	DataEncrypt   EventType = "data_encrypt"
	DataDecrypt   EventType = "data_decrypt"
	DataAccess    EventType = "data_access"
	
	// System events
	SystemStart     EventType = "system_start"
	SystemStop      EventType = "system_stop"
	SystemError     EventType = "system_error"
	ConfigChange    EventType = "config_change"
	
	// Security events
	SecurityViolation EventType = "security_violation"
	RateLimitHit      EventType = "rate_limit_hit"
	PermissionDenied  EventType = "permission_denied"
	IntrusionAttempt  EventType = "intrusion_attempt"
)

// Severity represents the severity level of an audit event
type Severity string

const (
	SeverityLow      Severity = "low"
	SeverityMedium   Severity = "medium"
	SeverityHigh     Severity = "high"
	SeverityCritical Severity = "critical"
)

// AuditEvent represents a security audit event
type AuditEvent struct {
	Timestamp   time.Time              `json:"timestamp"`
	EventType   EventType              `json:"event_type"`
	Severity    Severity               `json:"severity"`
	UserID      string                 `json:"user_id,omitempty"`
	Source      string                 `json:"source"`
	Target      string                 `json:"target,omitempty"`
	Action      string                 `json:"action"`
	Result      string                 `json:"result"`
	Message     string                 `json:"message"`
	Details     map[string]interface{} `json:"details,omitempty"`
	RemoteAddr  string                 `json:"remote_addr,omitempty"`
	UserAgent   string                 `json:"user_agent,omitempty"`
	SessionID   string                 `json:"session_id,omitempty"`
	RequestID   string                 `json:"request_id,omitempty"`
}

// Logger provides structured audit logging capabilities
type Logger struct {
	logger          *logrus.Logger
	outputFormat    string
	includeSensitive bool
	logFile         *os.File
	retention       time.Duration
}

// Config holds audit logger configuration
type Config struct {
	LogLevel         string        // debug, info, warn, error
	OutputFormat     string        // json, text
	LogFile          string        // Path to log file
	IncludeSensitive bool          // Whether to include sensitive information
	RetentionDays    int           // How many days to retain logs
	MaxFileSize      int64         // Maximum size per log file in bytes
	MaxFiles         int           // Maximum number of log files to keep
}

// NewLogger creates a new audit logger with the specified configuration
func NewLogger(config Config) (*Logger, error) {
	logger := logrus.New()
	
	// Set log level
	level, err := logrus.ParseLevel(config.LogLevel)
	if err != nil {
		return nil, fmt.Errorf("invalid log level: %w", err)
	}
	logger.SetLevel(level)

	// Set output format
	if config.OutputFormat == "json" {
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339Nano,
			PrettyPrint:     false,
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "timestamp",
				logrus.FieldKeyLevel: "level",
				logrus.FieldKeyMsg:   "message",
			},
		})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: time.RFC3339Nano,
			FullTimestamp:   true,
		})
	}

	auditLogger := &Logger{
		logger:           logger,
		outputFormat:     config.OutputFormat,
		includeSensitive: config.IncludeSensitive,
		retention:        time.Duration(config.RetentionDays) * 24 * time.Hour,
	}

	// Set up file output if specified
	if config.LogFile != "" {
		if err := auditLogger.setupFileOutput(config.LogFile); err != nil {
			return nil, fmt.Errorf("failed to setup file output: %w", err)
		}
	}

	return auditLogger, nil
}

// setupFileOutput configures file-based logging
func (l *Logger) setupFileOutput(logFile string) error {
	// Ensure directory exists
	dir := filepath.Dir(logFile)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Open log file with secure permissions
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0640)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	l.logFile = file
	l.logger.SetOutput(file)

	return nil
}

// LogEvent logs a security audit event
func (l *Logger) LogEvent(event AuditEvent) {
	// Set timestamp if not provided
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}

	// Sanitize sensitive information if needed
	if !l.includeSensitive {
		event = l.sanitizeEvent(event)
	}

	// Create logrus fields from the event
	fields := logrus.Fields{
		"audit_event_type": string(event.EventType),
		"severity":         string(event.Severity),
		"source":           event.Source,
		"action":           event.Action,
		"result":           event.Result,
	}
	
	// Override the timestamp with our event timestamp
	entry := l.logger.WithFields(fields).WithTime(event.Timestamp)

	// Add optional fields to the entry
	if event.UserID != "" {
		entry = entry.WithField("user_id", event.UserID)
	}
	if event.Target != "" {
		entry = entry.WithField("target", event.Target)
	}
	if event.RemoteAddr != "" {
		entry = entry.WithField("remote_addr", event.RemoteAddr)
	}
	if event.UserAgent != "" {
		entry = entry.WithField("user_agent", event.UserAgent)
	}
	if event.SessionID != "" {
		entry = entry.WithField("session_id", event.SessionID)
	}
	if event.RequestID != "" {
		entry = entry.WithField("request_id", event.RequestID)
	}

	// Add details if present
	if event.Details != nil {
		for key, value := range event.Details {
			entry = entry.WithField("detail_"+key, value)
		}
	}

	// Log based on severity
	switch event.Severity {
	case SeverityCritical:
		entry.Error(event.Message)
	case SeverityHigh:
		entry.Warn(event.Message)
	case SeverityMedium:
		entry.Info(event.Message)
	case SeverityLow:
		entry.Debug(event.Message)
	default:
		entry.Info(event.Message)
	}
}

// sanitizeEvent removes or masks sensitive information from events
func (l *Logger) sanitizeEvent(event AuditEvent) AuditEvent {
	// Create a copy to avoid modifying the original
	sanitized := event
	
	// Sanitize message
	sanitized.Message = l.sanitizeString(event.Message)
	
	// Sanitize details
	if event.Details != nil {
		sanitized.Details = make(map[string]interface{})
		for key, value := range event.Details {
			if l.isSensitiveField(key) {
				sanitized.Details[key] = "[REDACTED]"
			} else if strValue, ok := value.(string); ok {
				sanitized.Details[key] = l.sanitizeString(strValue)
			} else {
				sanitized.Details[key] = value
			}
		}
	}

	return sanitized
}

// sanitizeString masks sensitive patterns in strings
func (l *Logger) sanitizeString(s string) string {
	// List of sensitive patterns to mask
	sensitivePatterns := []string{
		"password", "secret", "token", "key", "credential",
		"client_secret", "access_token", "api_key",
	}

	result := s
	lower := strings.ToLower(result)
	
	for _, pattern := range sensitivePatterns {
		if strings.Contains(lower, pattern) {
			// Look for patterns like "password=value", "token: value", etc.
			// Replace everything after the = or : with [REDACTED]
			if idx := strings.Index(lower, pattern+"="); idx != -1 {
				start := idx + len(pattern) + 1
				end := len(result)
				// Find end of value (space or end of string)
				for i := start; i < len(result); i++ {
					if result[i] == ' ' || result[i] == '\n' || result[i] == '\t' {
						end = i
						break
					}
				}
				result = result[:start] + "[REDACTED]" + result[end:]
				lower = strings.ToLower(result) // Update lower case version
			} else if idx := strings.Index(lower, pattern+":"); idx != -1 {
				start := idx + len(pattern) + 1
				// Skip any spaces after the colon
				for start < len(result) && result[start] == ' ' {
					start++
				}
				end := len(result)
				// Find end of value (space or end of string)
				for i := start; i < len(result); i++ {
					if result[i] == ' ' || result[i] == '\n' || result[i] == '\t' {
						end = i
						break
					}
				}
				result = result[:start] + "[REDACTED]" + result[end:]
				lower = strings.ToLower(result) // Update lower case version
			}
		}
	}

	return result
}

// isSensitiveField checks if a field name indicates sensitive data
func (l *Logger) isSensitiveField(fieldName string) bool {
	sensitiveFields := []string{
		"password", "secret", "token", "key", "credential",
		"client_secret", "access_token", "api_key", "encryption_key",
	}

	fieldLower := strings.ToLower(fieldName)
	for _, sensitive := range sensitiveFields {
		if strings.Contains(fieldLower, sensitive) {
			return true
		}
	}

	return false
}

// LogCredentialAccess logs credential access events
func (l *Logger) LogCredentialAccess(userID, credentialType, result string) {
	severity := SeverityMedium
	if result != "success" {
		severity = SeverityHigh
	}

	event := AuditEvent{
		EventType: CredentialAccess,
		Severity:  severity,
		UserID:    userID,
		Source:    "credential_manager",
		Target:    credentialType,
		Action:    "access",
		Result:    result,
		Message:   fmt.Sprintf("Credential access attempt for %s: %s", credentialType, result),
		Details: map[string]interface{}{
			"credential_type": credentialType,
		},
	}

	l.LogEvent(event)
}

// LogAuthEvent logs authentication events
func (l *Logger) LogAuthEvent(eventType EventType, userID, result, remoteAddr string) {
	severity := SeverityMedium
	if result != "success" {
		severity = SeverityHigh
	}

	event := AuditEvent{
		EventType:  eventType,
		Severity:   severity,
		UserID:     userID,
		Source:     "auth_system",
		Action:     "authenticate",
		Result:     result,
		Message:    fmt.Sprintf("Authentication %s for user %s", result, userID),
		RemoteAddr: remoteAddr,
	}

	l.LogEvent(event)
}

// LogDataEvent logs data access and manipulation events
func (l *Logger) LogDataEvent(eventType EventType, userID, dataType, action, result string) {
	severity := SeverityMedium
	if strings.Contains(strings.ToLower(action), "export") {
		severity = SeverityHigh
	}

	event := AuditEvent{
		EventType: eventType,
		Severity:  severity,
		UserID:    userID,
		Source:    "data_system",
		Target:    dataType,
		Action:    action,
		Result:    result,
		Message:   fmt.Sprintf("Data %s on %s: %s", action, dataType, result),
		Details: map[string]interface{}{
			"data_type": dataType,
		},
	}

	l.LogEvent(event)
}

// LogSecurityViolation logs security violations
func (l *Logger) LogSecurityViolation(violation, source, description, remoteAddr string) {
	event := AuditEvent{
		EventType:  SecurityViolation,
		Severity:   SeverityCritical,
		Source:     source,
		Action:     violation,
		Result:     "blocked",
		Message:    fmt.Sprintf("Security violation detected: %s", violation),
		RemoteAddr: remoteAddr,
		Details: map[string]interface{}{
			"violation_type": violation,
			"description":    description,
		},
	}

	l.LogEvent(event)
}

// LogSystemEvent logs system-level events
func (l *Logger) LogSystemEvent(eventType EventType, component, action, result string) {
	severity := SeverityMedium
	if eventType == SystemError {
		severity = SeverityHigh
	}

	event := AuditEvent{
		EventType: eventType,
		Severity:  severity,
		Source:    component,
		Action:    action,
		Result:    result,
		Message:   fmt.Sprintf("System %s: %s", action, result),
		Details: map[string]interface{}{
			"component": component,
		},
	}

	l.LogEvent(event)
}

// CleanupOldLogs removes log files older than the retention period
func (l *Logger) CleanupOldLogs() error {
	if l.logFile == nil {
		return nil // No file logging configured
	}

	logDir := filepath.Dir(l.logFile.Name())
	
	return filepath.Walk(logDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check if file is a log file and older than retention period
		if strings.HasSuffix(path, ".log") {
			if time.Since(info.ModTime()) > l.retention {
				if err := os.Remove(path); err != nil {
					l.logger.WithError(err).Warnf("Failed to remove old log file: %s", path)
				} else {
					l.logger.Infof("Removed old log file: %s", path)
				}
			}
		}

		return nil
	})
}

// Close closes the audit logger and any open files
func (l *Logger) Close() error {
	if l.logFile != nil {
		err := l.logFile.Close()
		l.logFile = nil
		return err
	}
	return nil
}

// SetOutput sets a custom output writer for the logger
func (l *Logger) SetOutput(w io.Writer) {
	l.logger.SetOutput(w)
}

// GetMetrics returns audit logging metrics
func (l *Logger) GetMetrics() map[string]interface{} {
	return map[string]interface{}{
		"output_format":      l.outputFormat,
		"include_sensitive":  l.includeSensitive,
		"retention_hours":    l.retention.Hours(),
		"log_level":          l.logger.Level.String(),
	}
} 