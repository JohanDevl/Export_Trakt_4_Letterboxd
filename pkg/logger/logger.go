package logger

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

// Translator interface for i18n support
type Translator interface {
	Translate(messageID string, templateData map[string]interface{}) string
}

// Logger interface defines the logging methods
type Logger interface {
	Info(messageID string, data ...map[string]interface{})
	Infof(messageID string, data map[string]interface{})
	Error(messageID string, data ...map[string]interface{})
	Errorf(messageID string, data map[string]interface{})
	Warn(messageID string, data ...map[string]interface{})
	Warnf(messageID string, data map[string]interface{})
	Debug(messageID string, data ...map[string]interface{})
	Debugf(messageID string, data map[string]interface{})
	SetLogLevel(level string)
	SetLogFile(path string) error
	SetTranslator(t Translator)
}

// VisualFormatter provides a more readable, visual log format
type VisualFormatter struct {
	isQuietMode bool
}

// Format implements logrus.Formatter interface
func (f *VisualFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	timestamp := entry.Time.Format("15:04:05")
	
	var icon, levelStr string
	switch entry.Level {
	case logrus.ErrorLevel:
		icon = "❌"
		levelStr = "ERROR"
	case logrus.WarnLevel:
		icon = "⚠️ "
		levelStr = "WARN "
	case logrus.InfoLevel:
		// Special icons for specific messages
		if strings.Contains(entry.Message, "Successfully exported") {
			icon = "✅"
		} else if strings.Contains(entry.Message, "Retrieved") && strings.Contains(entry.Message, "movies") {
			icon = "📥"
		} else if strings.Contains(entry.Message, "Scheduler") {
			icon = "⏰"
		} else if strings.Contains(entry.Message, "Starting") || strings.Contains(entry.Message, "Initializing") {
			icon = "🚀"
		} else {
			icon = "ℹ️ "
		}
		levelStr = "INFO "
	case logrus.DebugLevel:
		icon = "🔧"
		levelStr = "DEBUG"
	default:
		icon = "📝"
		levelStr = "LOG  "
	}
	
	// In quiet mode, format important messages more prominently
	if f.isQuietMode {
		if strings.Contains(entry.Message, "Successfully exported") {
			return []byte(fmt.Sprintf("\n%s %s SUCCESS: %s\n", icon, timestamp, entry.Message)), nil
		} else if strings.Contains(entry.Message, "Retrieved") && strings.Contains(entry.Message, "movies") {
			return []byte(fmt.Sprintf("%s %s DATA: %s\n", icon, timestamp, entry.Message)), nil
		} else if strings.Contains(entry.Message, "Scheduler is running") {
			return []byte(fmt.Sprintf("%s %s STATUS: %s\n\n", icon, timestamp, entry.Message)), nil
		}
	}
	
	// Standard format
	return []byte(fmt.Sprintf("%s %s [%s] %s\n", icon, timestamp, levelStr, entry.Message)), nil
}

// DualWriter writes to both file and stdout, with filtering for stdout
type DualWriter struct {
	fileWriter   io.Writer
	stdoutWriter io.Writer
	quietMode    bool
}

// Write implements io.Writer interface
func (dw *DualWriter) Write(p []byte) (n int, err error) {
	// Always write to file
	if dw.fileWriter != nil {
		dw.fileWriter.Write(p)
	}
	
	// Filter stdout output in quiet mode
	if dw.quietMode {
		message := string(p)
		// Only show important messages in quiet mode
		if strings.Contains(message, "Successfully exported") ||
		   strings.Contains(message, "Scheduler is running") ||
		   strings.Contains(message, "Retrieved") ||
		   strings.Contains(message, "❌") || // Error icon
		   strings.Contains(message, "level=error") ||
		   strings.Contains(message, "level=fatal") {
			return dw.stdoutWriter.Write(p)
		}
		// Don't write to stdout for filtered messages
		return len(p), nil
	}
	
	// In non-quiet mode, write to stdout
	return dw.stdoutWriter.Write(p)
}

// StandardLogger wraps logrus.Logger
type StandardLogger struct {
	*logrus.Logger
	translator  Translator
	fileWriter  *os.File
	isQuietMode bool
}

// NewLogger creates a new logger instance
func NewLogger() Logger {
	log := logrus.New()
	log.SetOutput(os.Stdout)
	
	// Check if quiet mode is enabled
	quietMode := os.Getenv("EXPORT_QUIET_MODE") == "true"
	
	// Set visual formatter
	log.SetFormatter(&VisualFormatter{
		isQuietMode: quietMode,
	})
	log.SetLevel(logrus.InfoLevel)

	return &StandardLogger{
		Logger:      log,
		isQuietMode: quietMode,
	}
}

// SetTranslator sets the translator for the logger
func (l *StandardLogger) SetTranslator(t Translator) {
	l.translator = t
}

// translate handles message translation if a translator is available
func (l *StandardLogger) translate(messageID string, data map[string]interface{}) string {
	// No translation if no translator
	if l.translator == nil {
		return messageID
	}
	
	// Prevent recursion from specific error types
	if messageID == "" || messageID == "errors.translation_failed" {
		return messageID
	}
	
	// Sanitize the data to avoid nil map issues
	if data == nil {
		data = make(map[string]interface{})
	}
	
	return l.translator.Translate(messageID, data)
}

// Info logs an info level message with translation
func (l *StandardLogger) Info(messageID string, data ...map[string]interface{}) {
	var templateData map[string]interface{}
	if len(data) > 0 {
		templateData = data[0]
	}
	message := l.translate(messageID, templateData)
	
	// Enhance certain messages for better readability
	if strings.Contains(messageID, "scheduler.started") {
		message = "Scheduler started successfully!"
	} else if strings.Contains(messageID, "scheduler.waiting") {
		message = "Scheduler is running. Press Ctrl+C to stop..."
	}
	
	l.Logger.Info(message)
}

// Infof logs a formatted info level message with translation
func (l *StandardLogger) Infof(messageID string, data map[string]interface{}) {
	l.Logger.Info(l.translate(messageID, data))
}

// Error logs an error level message with translation
func (l *StandardLogger) Error(messageID string, data ...map[string]interface{}) {
	var templateData map[string]interface{}
	if len(data) > 0 {
		templateData = data[0]
	}
	l.Logger.Error(l.translate(messageID, templateData))
}

// Errorf logs a formatted error level message with translation
func (l *StandardLogger) Errorf(messageID string, data map[string]interface{}) {
	l.Logger.Error(l.translate(messageID, data))
}

// Warn logs a warning level message with translation
func (l *StandardLogger) Warn(messageID string, data ...map[string]interface{}) {
	var templateData map[string]interface{}
	if len(data) > 0 {
		templateData = data[0]
	}
	
	// Skip translation warnings in quiet mode to reduce noise
	if l.isQuietMode && strings.Contains(messageID, "translation_not_found") {
		return
	}
	
	l.Logger.Warn(l.translate(messageID, templateData))
}

// Warnf logs a formatted warning level message with translation
func (l *StandardLogger) Warnf(messageID string, data map[string]interface{}) {
	// Skip translation warnings in quiet mode to reduce noise
	if l.isQuietMode && strings.Contains(messageID, "translation_not_found") {
		return
	}
	l.Logger.Warn(l.translate(messageID, data))
}

// Debug logs a debug level message with translation
func (l *StandardLogger) Debug(messageID string, data ...map[string]interface{}) {
	var templateData map[string]interface{}
	if len(data) > 0 {
		templateData = data[0]
	}
	l.Logger.Debug(l.translate(messageID, templateData))
}

// Debugf logs a formatted debug level message with translation
func (l *StandardLogger) Debugf(messageID string, data map[string]interface{}) {
	l.Logger.Debug(l.translate(messageID, data))
}

// SetLogLevel sets the logging level
func (l *StandardLogger) SetLogLevel(level string) {
	switch level {
	case "debug":
		l.SetLevel(logrus.DebugLevel)
	case "info":
		l.SetLevel(logrus.InfoLevel)
	case "warn":
		l.SetLevel(logrus.WarnLevel)
	case "error":
		l.SetLevel(logrus.ErrorLevel)
	default:
		l.SetLevel(logrus.InfoLevel)
	}
}

// SetLogFile sets up dual output to both file and stdout
func (l *StandardLogger) SetLogFile(path string) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	
	l.fileWriter = file
	
	// Create dual writer
	dualWriter := &DualWriter{
		fileWriter:   file,
		stdoutWriter: os.Stdout,
		quietMode:    l.isQuietMode,
	}
	
	l.SetOutput(dualWriter)
	return nil
} 