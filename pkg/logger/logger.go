package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

// Translator interface for i18n support
type Translator interface {
	Translate(messageID string, templateData map[string]interface{}) string
}

// Logger wraps logrus.Logger
type Logger struct {
	*logrus.Logger
	translator Translator
}

// NewLogger creates a new logger instance
func NewLogger() *Logger {
	log := logrus.New()
	log.SetOutput(os.Stdout)
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	log.SetLevel(logrus.InfoLevel)

	return &Logger{
		Logger: log,
	}
}

// SetTranslator sets the translator for the logger
func (l *Logger) SetTranslator(t Translator) {
	l.translator = t
}

// translate handles message translation if a translator is available
func (l *Logger) translate(messageID string, data map[string]interface{}) string {
	if l.translator != nil {
		return l.translator.Translate(messageID, data)
	}
	return messageID
}

// Info logs an info level message with translation
func (l *Logger) Info(messageID string, data ...map[string]interface{}) {
	var templateData map[string]interface{}
	if len(data) > 0 {
		templateData = data[0]
	}
	l.Logger.Info(l.translate(messageID, templateData))
}

// Infof logs a formatted info level message with translation
func (l *Logger) Infof(messageID string, data map[string]interface{}) {
	l.Logger.Info(l.translate(messageID, data))
}

// Error logs an error level message with translation
func (l *Logger) Error(messageID string, data ...map[string]interface{}) {
	var templateData map[string]interface{}
	if len(data) > 0 {
		templateData = data[0]
	}
	l.Logger.Error(l.translate(messageID, templateData))
}

// Errorf logs a formatted error level message with translation
func (l *Logger) Errorf(messageID string, data map[string]interface{}) {
	l.Logger.Error(l.translate(messageID, data))
}

// Warn logs a warning level message with translation
func (l *Logger) Warn(messageID string, data ...map[string]interface{}) {
	var templateData map[string]interface{}
	if len(data) > 0 {
		templateData = data[0]
	}
	l.Logger.Warn(l.translate(messageID, templateData))
}

// Warnf logs a formatted warning level message with translation
func (l *Logger) Warnf(messageID string, data map[string]interface{}) {
	l.Logger.Warn(l.translate(messageID, data))
}

// Debug logs a debug level message with translation
func (l *Logger) Debug(messageID string, data ...map[string]interface{}) {
	var templateData map[string]interface{}
	if len(data) > 0 {
		templateData = data[0]
	}
	l.Logger.Debug(l.translate(messageID, templateData))
}

// Debugf logs a formatted debug level message with translation
func (l *Logger) Debugf(messageID string, data map[string]interface{}) {
	l.Logger.Debug(l.translate(messageID, data))
}

// SetLogLevel sets the logging level
func (l *Logger) SetLogLevel(level string) {
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
func (l *Logger) SetLogFile(path string) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	l.SetOutput(file)
	return nil
} 