package logger

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/monitoring"
)

// StructuredLogger extends the existing logger with enhanced capabilities
type StructuredLogger struct {
	*StandardLogger
	config monitoring.LoggingConfig
	hooks  []logrus.Hook
}

// CorrelationIDKey is the context key for correlation IDs
const CorrelationIDKey = "correlation_id"

// NewStructuredLogger creates a new structured logger with enhanced features
func NewStructuredLogger(config monitoring.LoggingConfig) (*StructuredLogger, error) {
	baseLogger := NewLogger().(*StandardLogger)
	
	sl := &StructuredLogger{
		StandardLogger: baseLogger,
		config:         config,
		hooks:          make([]logrus.Hook, 0),
	}

	// Configure output format
	if err := sl.configureFormatter(); err != nil {
		return nil, err
	}

	// Configure output destination
	if err := sl.configureOutput(); err != nil {
		return nil, err
	}

	// Set log level
	sl.configureLevel()

	// Add correlation ID hook if enabled
	if config.CorrelationID {
		sl.AddHook(&CorrelationIDHook{})
	}

	// Add tracing hook for integration with OpenTelemetry
	sl.AddHook(&TracingHook{})

	return sl, nil
}

// configureFormatter sets up the log formatter based on configuration
func (sl *StructuredLogger) configureFormatter() error {
	switch strings.ToLower(sl.config.Format) {
	case "json":
		sl.Logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "timestamp",
				logrus.FieldKeyLevel: "level",
				logrus.FieldKeyMsg:   "message",
				logrus.FieldKeyFunc:  "function",
			},
		})
	case "text":
		sl.Logger.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: time.RFC3339,
			FullTimestamp:   true,
		})
	default:
		// Keep the existing visual formatter
		sl.Logger.SetFormatter(&VisualFormatter{
			isQuietMode: sl.isQuietMode,
		})
	}
	return nil
}

// configureOutput sets up the log output destination
func (sl *StructuredLogger) configureOutput() error {
	switch strings.ToLower(sl.config.Output) {
	case "stdout":
		sl.Logger.SetOutput(os.Stdout)
	case "stderr":
		sl.Logger.SetOutput(os.Stderr)
	case "file":
		if sl.config.RotationEnabled {
			// TODO: Implement log rotation
			// For now, just use a simple file
			logDir := filepath.Dir("logs/app.log")
			if err := os.MkdirAll(logDir, 0755); err != nil {
				return err
			}
			
			file, err := os.OpenFile("logs/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
			if err != nil {
				return err
			}
			sl.Logger.SetOutput(file)
		}
	default:
		// Keep current output
	}
	return nil
}

// configureLevel sets the log level based on configuration
func (sl *StructuredLogger) configureLevel() {
	switch strings.ToLower(sl.config.Level) {
	case "debug":
		sl.Logger.SetLevel(logrus.DebugLevel)
	case "info":
		sl.Logger.SetLevel(logrus.InfoLevel)
	case "warn", "warning":
		sl.Logger.SetLevel(logrus.WarnLevel)
	case "error":
		sl.Logger.SetLevel(logrus.ErrorLevel)
	case "fatal":
		sl.Logger.SetLevel(logrus.FatalLevel)
	default:
		sl.Logger.SetLevel(logrus.InfoLevel)
	}
}

// AddHook adds a hook to the logger
func (sl *StructuredLogger) AddHook(hook logrus.Hook) {
	sl.Logger.AddHook(hook)
	sl.hooks = append(sl.hooks, hook)
}

// WithCorrelationID creates a context with a correlation ID
func (sl *StructuredLogger) WithCorrelationID(ctx context.Context) context.Context {
	correlationID := uuid.New().String()
	return context.WithValue(ctx, CorrelationIDKey, correlationID)
}

// WithFields creates a logger entry with structured fields
func (sl *StructuredLogger) WithFields(fields logrus.Fields) *logrus.Entry {
	return sl.Logger.WithFields(fields)
}

// WithField creates a logger entry with a single field
func (sl *StructuredLogger) WithField(key string, value interface{}) *logrus.Entry {
	return sl.Logger.WithField(key, value)
}

// WithContext creates a logger entry with context information
func (sl *StructuredLogger) WithContext(ctx context.Context) *logrus.Entry {
	entry := sl.Logger.WithContext(ctx)
	
	// Add correlation ID if present
	if correlationID := ctx.Value(CorrelationIDKey); correlationID != nil {
		entry = entry.WithField("correlation_id", correlationID)
	}
	
	// Add trace information if present
	if span := trace.SpanFromContext(ctx); span.SpanContext().IsValid() {
		entry = entry.WithFields(logrus.Fields{
			"trace_id": span.SpanContext().TraceID().String(),
			"span_id":  span.SpanContext().SpanID().String(),
		})
	}
	
	return entry
}

// InfoWithContext logs an info message with context
func (sl *StructuredLogger) InfoWithContext(ctx context.Context, messageID string, data ...map[string]interface{}) {
	var templateData map[string]interface{}
	if len(data) > 0 {
		templateData = data[0]
	}
	
	message := sl.translate(messageID, templateData)
	sl.WithContext(ctx).Info(message)
}

// ErrorWithContext logs an error message with context
func (sl *StructuredLogger) ErrorWithContext(ctx context.Context, messageID string, data ...map[string]interface{}) {
	var templateData map[string]interface{}
	if len(data) > 0 {
		templateData = data[0]
	}
	
	message := sl.translate(messageID, templateData)
	sl.WithContext(ctx).Error(message)
}

// WarnWithContext logs a warning message with context
func (sl *StructuredLogger) WarnWithContext(ctx context.Context, messageID string, data ...map[string]interface{}) {
	var templateData map[string]interface{}
	if len(data) > 0 {
		templateData = data[0]
	}
	
	message := sl.translate(messageID, templateData)
	sl.WithContext(ctx).Warn(message)
}

// DebugWithContext logs a debug message with context
func (sl *StructuredLogger) DebugWithContext(ctx context.Context, messageID string, data ...map[string]interface{}) {
	var templateData map[string]interface{}
	if len(data) > 0 {
		templateData = data[0]
	}
	
	message := sl.translate(messageID, templateData)
	sl.WithContext(ctx).Debug(message)
}

// CorrelationIDHook adds correlation IDs to log entries
type CorrelationIDHook struct{}

// Levels returns the log levels this hook should be fired for
func (hook *CorrelationIDHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire executes the hook
func (hook *CorrelationIDHook) Fire(entry *logrus.Entry) error {
	if entry.Context != nil {
		if correlationID := entry.Context.Value(CorrelationIDKey); correlationID != nil {
			entry.Data["correlation_id"] = correlationID
		}
	}
	return nil
}

// TracingHook integrates with OpenTelemetry tracing
type TracingHook struct{}

// Levels returns the log levels this hook should be fired for
func (hook *TracingHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire executes the hook
func (hook *TracingHook) Fire(entry *logrus.Entry) error {
	if entry.Context != nil {
		if span := trace.SpanFromContext(entry.Context); span.SpanContext().IsValid() {
			entry.Data["trace_id"] = span.SpanContext().TraceID().String()
			entry.Data["span_id"] = span.SpanContext().SpanID().String()
			
			// Add log entry as span event for important logs
			if entry.Level <= logrus.WarnLevel {
				span.AddEvent("log", trace.WithAttributes(
					attribute.String("level", entry.Level.String()),
					attribute.String("message", entry.Message),
				))
			}
		}
	}
	return nil
}

// SecurityHook sanitizes sensitive information from logs
type SecurityHook struct {
	sensitiveFields []string
}

// NewSecurityHook creates a new security hook
func NewSecurityHook(sensitiveFields []string) *SecurityHook {
	return &SecurityHook{
		sensitiveFields: sensitiveFields,
	}
}

// Levels returns the log levels this hook should be fired for
func (hook *SecurityHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire executes the hook to sanitize sensitive data
func (hook *SecurityHook) Fire(entry *logrus.Entry) error {
	for _, field := range hook.sensitiveFields {
		if _, exists := entry.Data[field]; exists {
			entry.Data[field] = "[REDACTED]"
		}
	}
	
	// Also sanitize in the message
	message := entry.Message
	for _, field := range hook.sensitiveFields {
		// Basic sanitization - replace common sensitive patterns
		if strings.Contains(strings.ToLower(message), strings.ToLower(field)) {
			// This is a simple approach - in production you'd want more sophisticated sanitization
			entry.Message = strings.ReplaceAll(message, field, "[REDACTED]")
		}
	}
	
	return nil
}

// PerformanceHook tracks logging performance
type PerformanceHook struct {
	slowLogThreshold time.Duration
}

// NewPerformanceHook creates a new performance tracking hook
func NewPerformanceHook(slowLogThreshold time.Duration) *PerformanceHook {
	return &PerformanceHook{
		slowLogThreshold: slowLogThreshold,
	}
}

// Levels returns the log levels this hook should be fired for
func (hook *PerformanceHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire executes the hook to track performance
func (hook *PerformanceHook) Fire(entry *logrus.Entry) error {
	start := time.Now()
	
	// The actual logging happens after this hook
	// We can't easily measure the exact logging time with this approach
	// This is more of a placeholder for a more sophisticated implementation
	
	duration := time.Since(start)
	if duration > hook.slowLogThreshold {
		// Could send a metric or alert about slow logging
		entry.Data["slow_log_duration"] = duration.String()
	}
	
	return nil
}

// AsyncLogger provides asynchronous logging capabilities
type AsyncLogger struct {
	*StructuredLogger
	logChan chan logrus.Entry
	done    chan bool
}

// NewAsyncLogger creates a new asynchronous logger
func NewAsyncLogger(config monitoring.LoggingConfig, bufferSize int) (*AsyncLogger, error) {
	structuredLogger, err := NewStructuredLogger(config)
	if err != nil {
		return nil, err
	}
	
	al := &AsyncLogger{
		StructuredLogger: structuredLogger,
		logChan:         make(chan logrus.Entry, bufferSize),
		done:            make(chan bool),
	}
	
	// Start the async logging goroutine
	go al.processLogs()
	
	return al, nil
}

// processLogs processes log entries asynchronously
func (al *AsyncLogger) processLogs() {
	for {
		select {
		case entry := <-al.logChan:
			// Process the log entry
			al.StructuredLogger.Logger.Log(entry.Level, entry.Message)
		case <-al.done:
			// Drain remaining logs before shutdown
			for len(al.logChan) > 0 {
				entry := <-al.logChan
				al.StructuredLogger.Logger.Log(entry.Level, entry.Message)
			}
			return
		}
	}
}

// AsyncInfo logs an info message asynchronously
func (al *AsyncLogger) AsyncInfo(message string, fields logrus.Fields) {
	entry := logrus.Entry{
		Logger:  al.StructuredLogger.Logger,
		Data:    fields,
		Time:    time.Now(),
		Level:   logrus.InfoLevel,
		Message: message,
	}
	
	select {
	case al.logChan <- entry:
	default:
		// Channel is full, fall back to synchronous logging
		al.StructuredLogger.Logger.WithFields(fields).Info(message)
	}
}

// Close shuts down the async logger
func (al *AsyncLogger) Close() {
	close(al.done)
} 