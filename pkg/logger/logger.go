package logger

import (
	"os"

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

// StandardLogger wraps logrus.Logger
type StandardLogger struct {
	*logrus.Logger
	translator Translator
}

// NewLogger creates a new logger instance
func NewLogger() Logger {
	log := logrus.New()
	log.SetOutput(os.Stdout)
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	log.SetLevel(logrus.InfoLevel)

	return &StandardLogger{
		Logger: log,
	}
}

// SetTranslator sets the translator for the logger
func (l *StandardLogger) SetTranslator(t Translator) {
	l.translator = t
}

// translate handles message translation if a translator is available
func (l *StandardLogger) translate(messageID string, data map[string]interface{}) string {
	if l.translator != nil {
		return l.translator.Translate(messageID, data)
	}
	return messageID
}

// Info logs an info level message with translation
func (l *StandardLogger) Info(messageID string, data ...map[string]interface{}) {
	var templateData map[string]interface{}
	if len(data) > 0 {
		templateData = data[0]
	}
	l.Logger.Info(l.translate(messageID, templateData))
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
	l.Logger.Warn(l.translate(messageID, templateData))
}

// Warnf logs a formatted warning level message with translation
func (l *StandardLogger) Warnf(messageID string, data map[string]interface{}) {
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

// SetLogFile sets the output to a file
func (l *StandardLogger) SetLogFile(path string) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	l.SetOutput(file)
	return nil
} 