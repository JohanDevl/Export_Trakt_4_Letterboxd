package types

import (
	"context"
	"fmt"
	"time"
)

// AppError represents the main error structure for the application
type AppError struct {
	Code     string       `json:"code"`
	Message  string       `json:"message"`
	Details  string       `json:"details,omitempty"`
	Cause    error        `json:"-"`
	Context  ErrorContext `json:"context,omitempty"`
	Time     time.Time    `json:"timestamp"`
}

// ErrorContext provides additional context for errors
type ErrorContext struct {
	Operation   string            `json:"operation"`
	RequestID   string            `json:"request_id,omitempty"`
	UserID      string            `json:"user_id,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	StackTrace  string            `json:"stack_trace,omitempty"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap allows error unwrapping for errors.Is and errors.As
func (e *AppError) Unwrap() error {
	return e.Cause
}

// WithContext adds context to an error
func (e *AppError) WithContext(ctx context.Context) *AppError {
	if requestID := ctx.Value("request_id"); requestID != nil {
		if id, ok := requestID.(string); ok {
			e.Context.RequestID = id
		}
	}
	if userID := ctx.Value("user_id"); userID != nil {
		if id, ok := userID.(string); ok {
			e.Context.UserID = id
		}
	}
	return e
}

// WithMetadata adds metadata to the error context
func (e *AppError) WithMetadata(key, value string) *AppError {
	if e.Context.Metadata == nil {
		e.Context.Metadata = make(map[string]string)
	}
	e.Context.Metadata[key] = value
	return e
}

// NewAppError creates a new application error
func NewAppError(code, message string, cause error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Cause:   cause,
		Time:    time.Now(),
		Context: ErrorContext{},
	}
}

// NewAppErrorWithOperation creates a new application error with operation context
func NewAppErrorWithOperation(code, message, operation string, cause error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Cause:   cause,
		Time:    time.Now(),
		Context: ErrorContext{
			Operation: operation,
		},
	}
} 