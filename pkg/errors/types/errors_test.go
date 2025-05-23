package types

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestNewAppError(t *testing.T) {
	err := NewAppError(ErrNetworkTimeout, "test message", nil)
	
	if err.Code != ErrNetworkTimeout {
		t.Errorf("Expected code %s, got %s", ErrNetworkTimeout, err.Code)
	}
	
	if err.Message != "test message" {
		t.Errorf("Expected message 'test message', got %s", err.Message)
	}
	
	if err.Cause != nil {
		t.Errorf("Expected nil cause, got %v", err.Cause)
	}
}

func TestNewAppErrorWithOperation(t *testing.T) {
	cause := errors.New("original error")
	err := NewAppErrorWithOperation(ErrNetworkTimeout, "test message", "test_op", cause)
	
	if err.Context.Operation != "test_op" {
		t.Errorf("Expected operation 'test_op', got %s", err.Context.Operation)
	}
	
	if err.Cause != cause {
		t.Errorf("Expected cause to be set")
	}
}

func TestAppErrorWithContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), "request_id", "test-123")
	err := NewAppError(ErrNetworkTimeout, "test message", nil).WithContext(ctx)
	
	if err.Context.RequestID != "test-123" {
		t.Errorf("Expected request ID to be set to 'test-123', got %s", err.Context.RequestID)
	}
}

func TestAppErrorWithMetadata(t *testing.T) {
	err := NewAppError(ErrNetworkTimeout, "test message", nil).WithMetadata("key", "value")
	
	if err.Context.Metadata["key"] != "value" {
		t.Errorf("Expected metadata key=value, got %v", err.Context.Metadata["key"])
	}
}

func TestAppErrorError(t *testing.T) {
	cause := errors.New("original error")
	err := NewAppErrorWithOperation(ErrNetworkTimeout, "test message", "test_op", cause)
	
	expected := "NET_001: test message (caused by: original error)"
	if err.Error() != expected {
		t.Errorf("Expected error string '%s', got '%s'", expected, err.Error())
	}
}

func TestAppErrorErrorWithoutCause(t *testing.T) {
	err := NewAppError(ErrNetworkTimeout, "test message", nil)
	
	expected := "NET_001: test message"
	if err.Error() != expected {
		t.Errorf("Expected error string '%s', got '%s'", expected, err.Error())
	}
}

func TestAppErrorUnwrap(t *testing.T) {
	cause := errors.New("original error")
	err := NewAppError(ErrNetworkTimeout, "test message", cause)
	
	if !errors.Is(err, cause) {
		t.Errorf("Expected error to wrap original cause")
	}
}

func TestGetErrorCategory(t *testing.T) {
	tests := []struct {
		code     string
		category ErrorCategory
	}{
		{ErrNetworkTimeout, CategoryNetwork},
		{ErrInvalidCredentials, CategoryAuthentication},
		{ErrInvalidInput, CategoryValidation},
		{ErrExportFailed, CategoryOperation},
		{ErrDataCorrupted, CategoryData},
		{ErrConfigMissing, CategoryConfiguration},
		{ErrSystemResource, CategorySystem},
	}
	
	for _, test := range tests {
		category := GetErrorCategory(test.code)
		if category != test.category {
			t.Errorf("Expected category %s for code %s, got %s", test.category, test.code, category)
		}
	}
}

func TestIsRetryableError(t *testing.T) {
	retryableCodes := []string{
		ErrNetworkTimeout,
		ErrNetworkUnavailable,
		ErrRateLimited,
		ErrConnectionRefused,
		ErrTokenExpired,
		ErrSystemResource,
	}
	
	nonRetryableCodes := []string{
		ErrInvalidCredentials,
		ErrInvalidInput,
		ErrDataCorrupted,
		ErrSystemDisk,
		ErrSystemMemory,
	}
	
	for _, code := range retryableCodes {
		if !IsRetryableError(code) {
			t.Errorf("Expected code %s to be retryable", code)
		}
	}
	
	for _, code := range nonRetryableCodes {
		if IsRetryableError(code) {
			t.Errorf("Expected code %s to not be retryable", code)
		}
	}
}

func TestAppErrorTimestamp(t *testing.T) {
	before := time.Now()
	err := NewAppError(ErrNetworkTimeout, "test message", nil)
	after := time.Now()
	
	if err.Time.Before(before) || err.Time.After(after) {
		t.Errorf("Expected timestamp to be between %v and %v, got %v", before, after, err.Time)
	}
}

func TestAppErrorMetadataChaining(t *testing.T) {
	err := NewAppError(ErrNetworkTimeout, "test message", nil).
		WithMetadata("key1", "value1").
		WithMetadata("key2", "value2")
	
	if err.Context.Metadata["key1"] != "value1" {
		t.Errorf("Expected metadata key1=value1")
	}
	
	if err.Context.Metadata["key2"] != "value2" {
		t.Errorf("Expected metadata key2=value2")
	}
}

func TestIsTemporaryError(t *testing.T) {
	temporaryCodes := []string{
		ErrNetworkTimeout,
		ErrNetworkUnavailable,
		ErrRateLimited,
		ErrSystemResource,
		ErrTokenExpired,
	}
	
	nonTemporaryCodes := []string{
		ErrInvalidCredentials,
		ErrDataCorrupted,
		ErrConfigMissing,
	}
	
	for _, code := range temporaryCodes {
		if !IsTemporaryError(code) {
			t.Errorf("Expected code %s to be temporary", code)
		}
	}
	
	for _, code := range nonTemporaryCodes {
		if IsTemporaryError(code) {
			t.Errorf("Expected code %s to not be temporary", code)
		}
	}
} 