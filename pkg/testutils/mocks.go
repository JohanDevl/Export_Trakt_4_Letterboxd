package testutils

import (
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/logger"
)

// MockLogger provides a testify-based mock implementation of logger.Logger
type MockLogger struct {
	mock.Mock
	LastMessage string
	LastData    map[string]interface{}
}

func NewMockLogger() *MockLogger {
	return &MockLogger{
		LastData: make(map[string]interface{}),
	}
}

func (m *MockLogger) Info(messageID string, data ...map[string]interface{}) {
	m.LastMessage = messageID
	if len(data) > 0 {
		m.LastData = data[0]
	}
	m.Called(messageID, data)
}

func (m *MockLogger) Infof(messageID string, data map[string]interface{}) {
	m.LastMessage = messageID
	m.LastData = data
	m.Called(messageID, data)
}

func (m *MockLogger) Error(messageID string, data ...map[string]interface{}) {
	m.LastMessage = messageID
	if len(data) > 0 {
		m.LastData = data[0]
	}
	m.Called(messageID, data)
}

func (m *MockLogger) Errorf(messageID string, data map[string]interface{}) {
	m.LastMessage = messageID
	m.LastData = data
	m.Called(messageID, data)
}

func (m *MockLogger) Warn(messageID string, data ...map[string]interface{}) {
	m.LastMessage = messageID
	if len(data) > 0 {
		m.LastData = data[0]
	}
	m.Called(messageID, data)
}

func (m *MockLogger) Warnf(messageID string, data map[string]interface{}) {
	m.LastMessage = messageID
	m.LastData = data
	m.Called(messageID, data)
}

func (m *MockLogger) Debug(messageID string, data ...map[string]interface{}) {
	m.LastMessage = messageID
	if len(data) > 0 {
		m.LastData = data[0]
	}
	m.Called(messageID, data)
}

func (m *MockLogger) Debugf(messageID string, data map[string]interface{}) {
	m.LastMessage = messageID
	m.LastData = data
	m.Called(messageID, data)
}

func (m *MockLogger) SetLogLevel(level string) {
	m.Called(level)
}

func (m *MockLogger) SetLogFile(path string) error {
	args := m.Called(path)
	return args.Error(0)
}

func (m *MockLogger) SetTranslator(t logger.Translator) {
	m.Called(t)
}

// NoOpLogger provides a simple no-operation logger for tests that don't need to verify log calls
type NoOpLogger struct{}

func NewNoOpLogger() *NoOpLogger {
	return &NoOpLogger{}
}

func (l *NoOpLogger) Info(messageID string, data ...map[string]interface{})  {}
func (l *NoOpLogger) Infof(messageID string, data map[string]interface{})   {}
func (l *NoOpLogger) Error(messageID string, data ...map[string]interface{}) {}
func (l *NoOpLogger) Errorf(messageID string, data map[string]interface{})  {}
func (l *NoOpLogger) Warn(messageID string, data ...map[string]interface{})  {}
func (l *NoOpLogger) Warnf(messageID string, data map[string]interface{})   {}
func (l *NoOpLogger) Debug(messageID string, data ...map[string]interface{}) {}
func (l *NoOpLogger) Debugf(messageID string, data map[string]interface{})  {}
func (l *NoOpLogger) SetLogLevel(level string)                              {}
func (l *NoOpLogger) SetLogFile(path string) error                          { return nil }
func (l *NoOpLogger) SetTranslator(t logger.Translator)                     {}

// CapturingLogger captures log messages for inspection in tests
type CapturingLogger struct {
	Messages []LogMessage
}

type LogMessage struct {
	Level     string
	MessageID string
	Data      map[string]interface{}
}

func NewCapturingLogger() *CapturingLogger {
	return &CapturingLogger{
		Messages: make([]LogMessage, 0),
	}
}

func (l *CapturingLogger) Info(messageID string, data ...map[string]interface{}) {
	msg := LogMessage{Level: "info", MessageID: messageID}
	if len(data) > 0 {
		msg.Data = data[0]
	}
	l.Messages = append(l.Messages, msg)
}

func (l *CapturingLogger) Infof(messageID string, data map[string]interface{}) {
	l.Messages = append(l.Messages, LogMessage{Level: "info", MessageID: messageID, Data: data})
}

func (l *CapturingLogger) Error(messageID string, data ...map[string]interface{}) {
	msg := LogMessage{Level: "error", MessageID: messageID}
	if len(data) > 0 {
		msg.Data = data[0]
	}
	l.Messages = append(l.Messages, msg)
}

func (l *CapturingLogger) Errorf(messageID string, data map[string]interface{}) {
	l.Messages = append(l.Messages, LogMessage{Level: "error", MessageID: messageID, Data: data})
}

func (l *CapturingLogger) Warn(messageID string, data ...map[string]interface{}) {
	msg := LogMessage{Level: "warn", MessageID: messageID}
	if len(data) > 0 {
		msg.Data = data[0]
	}
	l.Messages = append(l.Messages, msg)
}

func (l *CapturingLogger) Warnf(messageID string, data map[string]interface{}) {
	l.Messages = append(l.Messages, LogMessage{Level: "warn", MessageID: messageID, Data: data})
}

func (l *CapturingLogger) Debug(messageID string, data ...map[string]interface{}) {
	msg := LogMessage{Level: "debug", MessageID: messageID}
	if len(data) > 0 {
		msg.Data = data[0]
	}
	l.Messages = append(l.Messages, msg)
}

func (l *CapturingLogger) Debugf(messageID string, data map[string]interface{}) {
	l.Messages = append(l.Messages, LogMessage{Level: "debug", MessageID: messageID, Data: data})
}

func (l *CapturingLogger) SetLogLevel(level string) {}

func (l *CapturingLogger) SetLogFile(path string) error { return nil }

func (l *CapturingLogger) SetTranslator(t logger.Translator) {}

// GetMessages returns all captured messages
func (l *CapturingLogger) GetMessages() []LogMessage {
	return l.Messages
}

// GetMessagesByLevel returns messages filtered by level
func (l *CapturingLogger) GetMessagesByLevel(level string) []LogMessage {
	var filtered []LogMessage
	for _, msg := range l.Messages {
		if msg.Level == level {
			filtered = append(filtered, msg)
		}
	}
	return filtered
}

// Clear removes all captured messages
func (l *CapturingLogger) Clear() {
	l.Messages = l.Messages[:0]
}

// MockTokenManager provides a mock implementation for token management
type MockTokenManager struct {
	mock.Mock
	Token string
	Err   error
}

func NewMockTokenManager() *MockTokenManager {
	return &MockTokenManager{
		Token: "test_token",
	}
}

func (m *MockTokenManager) GetValidAccessToken() (string, error) {
	if m.Err != nil {
		return "", m.Err
	}
	if m.Token != "" {
		return m.Token, nil
	}
	// Only call mock if expectations are set
	if len(m.ExpectedCalls) > 0 {
		args := m.Called()
		return args.String(0), args.Error(1)
	}
	return "", nil
}

// SetToken sets the token to return
func (m *MockTokenManager) SetToken(token string) {
	m.Token = token
}

// SetError sets an error to return
func (m *MockTokenManager) SetError(err error) {
	m.Err = err
}

// MockTranslator provides a mock implementation for translation
type MockTranslator struct {
	mock.Mock
	Translations map[string]string
}

func NewMockTranslator() *MockTranslator {
	return &MockTranslator{
		Translations: make(map[string]string),
	}
}

func (m *MockTranslator) Translate(messageID string, templateData map[string]interface{}) string {
	if translation, exists := m.Translations[messageID]; exists {
		return translation
	}
	// Only call mock if expectations are set
	if len(m.ExpectedCalls) > 0 {
		args := m.Called(messageID, templateData)
		if args.String(0) != "" {
			return args.String(0)
		}
	}
	return messageID // fallback to messageID if no translation
}

// SetTranslation adds a translation
func (m *MockTranslator) SetTranslation(messageID, translation string) {
	m.Translations[messageID] = translation
}

// MockMetricsRecorder provides a mock implementation for metrics recording
type MockMetricsRecorder struct {
	mock.Mock
	JobsProcessed int64
	JobsErrored   int64
	JobDurations  []time.Duration
}

func NewMockMetricsRecorder() *MockMetricsRecorder {
	return &MockMetricsRecorder{
		JobDurations: make([]time.Duration, 0),
	}
}

func (m *MockMetricsRecorder) IncrementJobsProcessed() {
	m.JobsProcessed++
	m.Called()
}

func (m *MockMetricsRecorder) RecordJobDuration(duration time.Duration) {
	m.JobDurations = append(m.JobDurations, duration)
	m.Called(duration)
}

func (m *MockMetricsRecorder) IncrementJobsErrored() {
	m.JobsErrored++
	m.Called()
}

// GetJobsProcessed returns the number of jobs processed
func (m *MockMetricsRecorder) GetJobsProcessed() int64 {
	return m.JobsProcessed
}

// GetJobsErrored returns the number of jobs errored
func (m *MockMetricsRecorder) GetJobsErrored() int64 {
	return m.JobsErrored
}

// GetJobDurations returns all recorded job durations
func (m *MockMetricsRecorder) GetJobDurations() []time.Duration {
	return m.JobDurations
}

// Reset clears all recorded metrics
func (m *MockMetricsRecorder) Reset() {
	m.JobsProcessed = 0
	m.JobsErrored = 0
	m.JobDurations = m.JobDurations[:0]
}