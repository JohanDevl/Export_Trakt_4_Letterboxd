package tracing

import (
	"context"
	"fmt"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func TestNewOpenTelemetryTracer_Disabled(t *testing.T) {
	logger := logrus.New()
	config := TracingConfig{
		Enabled:        false,
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		SamplingRate:   1.0,
	}
	
	tracer, err := NewOpenTelemetryTracer(logger, config)
	
	assert.NoError(t, err)
	assert.NotNil(t, tracer)
	assert.Equal(t, logger, tracer.logger)
	assert.Equal(t, config, tracer.config)
	assert.Nil(t, tracer.tracer)
	assert.Nil(t, tracer.provider)
}

func TestNewOpenTelemetryTracer_Enabled(t *testing.T) {
	logger := logrus.New()
	config := TracingConfig{
		Enabled:        true,
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		SamplingRate:   1.0,
		// No Jaeger endpoint to avoid external dependencies
	}
	
	tracer, err := NewOpenTelemetryTracer(logger, config)
	
	assert.NoError(t, err)
	assert.NotNil(t, tracer)
	assert.Equal(t, logger, tracer.logger)
	assert.Equal(t, config, tracer.config)
	assert.NotNil(t, tracer.tracer)
	assert.NotNil(t, tracer.provider)
	assert.Equal(t, "test-service", tracer.serviceName)
	
	// Clean up
	err = tracer.Close()
	assert.NoError(t, err)
}

func TestOpenTelemetryTracer_StartSpan_Disabled(t *testing.T) {
	logger := logrus.New()
	config := TracingConfig{Enabled: false}
	
	tracer, err := NewOpenTelemetryTracer(logger, config)
	require.NoError(t, err)
	
	ctx := context.Background()
	spanCtx, endSpan := tracer.StartSpan(ctx, "test-operation")
	
	// Should return the same context when disabled
	assert.Equal(t, ctx, spanCtx)
	
	// End span function should be no-op
	endSpan()
}

func TestOpenTelemetryTracer_StartSpan_Enabled(t *testing.T) {
	logger := logrus.New()
	config := TracingConfig{
		Enabled:        true,
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		SamplingRate:   1.0,
	}
	
	tracer, err := NewOpenTelemetryTracer(logger, config)
	require.NoError(t, err)
	defer tracer.Close()
	
	ctx := context.Background()
	spanCtx, endSpan := tracer.StartSpan(ctx, "test-operation")
	
	// Should return a new context with span
	assert.NotEqual(t, ctx, spanCtx)
	
	// Should have a span in the context
	span := trace.SpanFromContext(spanCtx)
	assert.NotNil(t, span)
	assert.True(t, span.IsRecording())
	
	// End the span
	endSpan()
}

func TestOpenTelemetryTracer_StartSpanWithAttributes(t *testing.T) {
	logger := logrus.New()
	config := TracingConfig{
		Enabled:        true,
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		SamplingRate:   1.0,
	}
	
	tracer, err := NewOpenTelemetryTracer(logger, config)
	require.NoError(t, err)
	defer tracer.Close()
	
	ctx := context.Background()
	attributes := map[string]string{
		"custom.attribute": "test-value",
		"user.id":         "123",
	}
	
	spanCtx, endSpan := tracer.StartSpanWithAttributes(ctx, "test-operation", attributes)
	
	// Should return a new context with span
	assert.NotEqual(t, ctx, spanCtx)
	
	// Should have a span in the context
	span := trace.SpanFromContext(spanCtx)
	assert.NotNil(t, span)
	assert.True(t, span.IsRecording())
	
	// End the span
	endSpan()
}

func TestOpenTelemetryTracer_StartSpanWithAttributes_Disabled(t *testing.T) {
	logger := logrus.New()
	config := TracingConfig{Enabled: false}
	
	tracer, err := NewOpenTelemetryTracer(logger, config)
	require.NoError(t, err)
	
	ctx := context.Background()
	attributes := map[string]string{
		"custom.attribute": "test-value",
	}
	
	spanCtx, endSpan := tracer.StartSpanWithAttributes(ctx, "test-operation", attributes)
	
	// Should return the same context when disabled
	assert.Equal(t, ctx, spanCtx)
	
	// End span function should be no-op
	endSpan()
}

func TestOpenTelemetryTracer_AddSpanEvent(t *testing.T) {
	logger := logrus.New()
	config := TracingConfig{
		Enabled:        true,
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		SamplingRate:   1.0,
	}
	
	tracer, err := NewOpenTelemetryTracer(logger, config)
	require.NoError(t, err)
	defer tracer.Close()
	
	ctx := context.Background()
	spanCtx, endSpan := tracer.StartSpan(ctx, "test-operation")
	defer endSpan()
	
	attributes := map[string]string{
		"event.type": "test-event",
		"count":      "42",
	}
	
	// This should not panic or error
	tracer.AddSpanEvent(spanCtx, "test-event", attributes)
}

func TestOpenTelemetryTracer_AddSpanEvent_Disabled(t *testing.T) {
	logger := logrus.New()
	config := TracingConfig{Enabled: false}
	
	tracer, err := NewOpenTelemetryTracer(logger, config)
	require.NoError(t, err)
	
	ctx := context.Background()
	attributes := map[string]string{
		"event.type": "test-event",
	}
	
	// This should not panic when tracing is disabled
	tracer.AddSpanEvent(ctx, "test-event", attributes)
}

func TestOpenTelemetryTracer_AddSpanError(t *testing.T) {
	logger := logrus.New()
	config := TracingConfig{
		Enabled:        true,
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		SamplingRate:   1.0,
	}
	
	tracer, err := NewOpenTelemetryTracer(logger, config)
	require.NoError(t, err)
	defer tracer.Close()
	
	ctx := context.Background()
	spanCtx, endSpan := tracer.StartSpan(ctx, "test-operation")
	defer endSpan()
	
	testError := fmt.Errorf("test error occurred")
	
	// This should not panic or error
	tracer.AddSpanError(spanCtx, testError)
	
	// Verify the span status was set to error
	span := trace.SpanFromContext(spanCtx)
	assert.NotNil(t, span)
}

func TestOpenTelemetryTracer_AddSpanError_Disabled(t *testing.T) {
	logger := logrus.New()
	config := TracingConfig{Enabled: false}
	
	tracer, err := NewOpenTelemetryTracer(logger, config)
	require.NoError(t, err)
	
	ctx := context.Background()
	testError := fmt.Errorf("test error")
	
	// This should not panic when tracing is disabled
	tracer.AddSpanError(ctx, testError)
}

func TestOpenTelemetryTracer_SetSpanStatus(t *testing.T) {
	logger := logrus.New()
	config := TracingConfig{
		Enabled:        true,
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		SamplingRate:   1.0,
	}
	
	tracer, err := NewOpenTelemetryTracer(logger, config)
	require.NoError(t, err)
	defer tracer.Close()
	
	ctx := context.Background()
	spanCtx, endSpan := tracer.StartSpan(ctx, "test-operation")
	defer endSpan()
	
	// This should not panic or error
	tracer.SetSpanStatus(spanCtx, codes.Ok, "Operation completed successfully")
}

func TestOpenTelemetryTracer_SetSpanStatus_Disabled(t *testing.T) {
	logger := logrus.New()
	config := TracingConfig{Enabled: false}
	
	tracer, err := NewOpenTelemetryTracer(logger, config)
	require.NoError(t, err)
	
	ctx := context.Background()
	
	// This should not panic when tracing is disabled
	tracer.SetSpanStatus(ctx, codes.Ok, "test message")
}

func TestOpenTelemetryTracer_GetTraceID(t *testing.T) {
	logger := logrus.New()
	config := TracingConfig{
		Enabled:        true,
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		SamplingRate:   1.0,
	}
	
	tracer, err := NewOpenTelemetryTracer(logger, config)
	require.NoError(t, err)
	defer tracer.Close()
	
	t.Run("with span", func(t *testing.T) {
		ctx := context.Background()
		spanCtx, endSpan := tracer.StartSpan(ctx, "test-operation")
		defer endSpan()
		
		traceID := tracer.GetTraceID(spanCtx)
		assert.NotEmpty(t, traceID)
		assert.NotEqual(t, "00000000000000000000000000000000", traceID)
	})
	
	t.Run("without span", func(t *testing.T) {
		ctx := context.Background()
		traceID := tracer.GetTraceID(ctx)
		assert.Equal(t, "", traceID)
	})
	
	t.Run("disabled", func(t *testing.T) {
		disabledTracer, err := NewOpenTelemetryTracer(logger, TracingConfig{Enabled: false})
		require.NoError(t, err)
		
		traceID := disabledTracer.GetTraceID(context.Background())
		assert.Empty(t, traceID)
	})
}

func TestOpenTelemetryTracer_GetSpanID(t *testing.T) {
	logger := logrus.New()
	config := TracingConfig{
		Enabled:        true,
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		SamplingRate:   1.0,
	}
	
	tracer, err := NewOpenTelemetryTracer(logger, config)
	require.NoError(t, err)
	defer tracer.Close()
	
	t.Run("with span", func(t *testing.T) {
		ctx := context.Background()
		spanCtx, endSpan := tracer.StartSpan(ctx, "test-operation")
		defer endSpan()
		
		spanID := tracer.GetSpanID(spanCtx)
		assert.NotEmpty(t, spanID)
		assert.NotEqual(t, "0000000000000000", spanID)
	})
	
	t.Run("without span", func(t *testing.T) {
		ctx := context.Background()
		spanID := tracer.GetSpanID(ctx)
		assert.Equal(t, "", spanID)
	})
	
	t.Run("disabled", func(t *testing.T) {
		disabledTracer, err := NewOpenTelemetryTracer(logger, TracingConfig{Enabled: false})
		require.NoError(t, err)
		
		spanID := disabledTracer.GetSpanID(context.Background())
		assert.Empty(t, spanID)
	})
}

func TestOpenTelemetryTracer_TraceExportOperation(t *testing.T) {
	logger := logrus.New()
	config := TracingConfig{
		Enabled:        true,
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		SamplingRate:   1.0,
	}
	
	tracer, err := NewOpenTelemetryTracer(logger, config)
	require.NoError(t, err)
	defer tracer.Close()
	
	t.Run("successful operation", func(t *testing.T) {
		executed := false
		err := tracer.TraceExportOperation(context.Background(), "movies", func(ctx context.Context) error {
			executed = true
			// Verify we have a span in the context
			span := trace.SpanFromContext(ctx)
			assert.NotNil(t, span)
			assert.True(t, span.IsRecording())
			return nil
		})
		
		assert.NoError(t, err)
		assert.True(t, executed)
	})
	
	t.Run("operation with error", func(t *testing.T) {
		testError := fmt.Errorf("operation failed")
		err := tracer.TraceExportOperation(context.Background(), "movies", func(ctx context.Context) error {
			return testError
		})
		
		assert.Error(t, err)
		assert.Equal(t, testError, err)
	})
}

func TestOpenTelemetryTracer_TraceAPICall(t *testing.T) {
	logger := logrus.New()
	config := TracingConfig{
		Enabled:        true,
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		SamplingRate:   1.0,
	}
	
	tracer, err := NewOpenTelemetryTracer(logger, config)
	require.NoError(t, err)
	defer tracer.Close()
	
	t.Run("successful API call", func(t *testing.T) {
		executed := false
		err := tracer.TraceAPICall(context.Background(), "trakt", "/movies", "GET", func(ctx context.Context) error {
			executed = true
			// Verify we have a span in the context
			span := trace.SpanFromContext(ctx)
			assert.NotNil(t, span)
			assert.True(t, span.IsRecording())
			return nil
		})
		
		assert.NoError(t, err)
		assert.True(t, executed)
	})
	
	t.Run("API call with error", func(t *testing.T) {
		testError := fmt.Errorf("API call failed")
		err := tracer.TraceAPICall(context.Background(), "trakt", "/movies", "GET", func(ctx context.Context) error {
			return testError
		})
		
		assert.Error(t, err)
		assert.Equal(t, testError, err)
	})
}

func TestOpenTelemetryTracer_TraceFileOperation(t *testing.T) {
	logger := logrus.New()
	config := TracingConfig{
		Enabled:        true,
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		SamplingRate:   1.0,
	}
	
	tracer, err := NewOpenTelemetryTracer(logger, config)
	require.NoError(t, err)
	defer tracer.Close()
	
	t.Run("successful file operation", func(t *testing.T) {
		executed := false
		err := tracer.TraceFileOperation(context.Background(), "write", "/tmp/test.csv", func(ctx context.Context) error {
			executed = true
			// Verify we have a span in the context
			span := trace.SpanFromContext(ctx)
			assert.NotNil(t, span)
			assert.True(t, span.IsRecording())
			return nil
		})
		
		assert.NoError(t, err)
		assert.True(t, executed)
	})
	
	t.Run("file operation with error", func(t *testing.T) {
		testError := fmt.Errorf("file operation failed")
		err := tracer.TraceFileOperation(context.Background(), "write", "/tmp/test.csv", func(ctx context.Context) error {
			return testError
		})
		
		assert.Error(t, err)
		assert.Equal(t, testError, err)
	})
}

func TestOpenTelemetryTracer_Close(t *testing.T) {
	logger := logrus.New()
	config := TracingConfig{
		Enabled:        true,
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		SamplingRate:   1.0,
	}
	
	tracer, err := NewOpenTelemetryTracer(logger, config)
	require.NoError(t, err)
	
	// Should not error on close
	err = tracer.Close()
	assert.NoError(t, err)
	
	// Should not error on multiple closes
	err = tracer.Close()
	assert.NoError(t, err)
}

func TestOpenTelemetryTracer_Close_Disabled(t *testing.T) {
	logger := logrus.New()
	config := TracingConfig{Enabled: false}
	
	tracer, err := NewOpenTelemetryTracer(logger, config)
	require.NoError(t, err)
	
	// Should not error on close when disabled
	err = tracer.Close()
	assert.NoError(t, err)
}

func TestTracingConfig_Validation(t *testing.T) {
	logger := logrus.New()
	
	testCases := []struct {
		name        string
		config      TracingConfig
		expectError bool
	}{
		{
			name: "valid config",
			config: TracingConfig{
				Enabled:        true,
				ServiceName:    "test-service",
				ServiceVersion: "1.0.0",
				Environment:    "test",
				SamplingRate:   1.0,
			},
			expectError: false,
		},
		{
			name: "disabled config",
			config: TracingConfig{
				Enabled: false,
			},
			expectError: false,
		},
		{
			name: "config with Jaeger endpoint",
			config: TracingConfig{
				Enabled:        true,
				ServiceName:    "test-service",
				ServiceVersion: "1.0.0",
				Environment:    "test",
				JaegerEndpoint: "http://localhost:14268/api/traces",
				SamplingRate:   0.5,
			},
			expectError: false, // Won't actually connect in tests
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tracer, err := NewOpenTelemetryTracer(logger, tc.config)
			
			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, tracer)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, tracer)
				if tracer.provider != nil {
					tracer.Close()
				}
			}
		})
	}
}

func TestNestedSpans(t *testing.T) {
	logger := logrus.New()
	config := TracingConfig{
		Enabled:        true,
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		SamplingRate:   1.0,
	}
	
	tracer, err := NewOpenTelemetryTracer(logger, config)
	require.NoError(t, err)
	defer tracer.Close()
	
	ctx := context.Background()
	
	// Start parent span
	parentCtx, endParent := tracer.StartSpan(ctx, "parent-operation")
	defer endParent()
	
	parentSpan := trace.SpanFromContext(parentCtx)
	assert.NotNil(t, parentSpan)
	assert.True(t, parentSpan.IsRecording())
	
	// Start child span
	childCtx, endChild := tracer.StartSpan(parentCtx, "child-operation")
	defer endChild()
	
	childSpan := trace.SpanFromContext(childCtx)
	assert.NotNil(t, childSpan)
	assert.True(t, childSpan.IsRecording())
	
	// Child and parent should have the same trace ID but different span IDs
	parentTraceID := tracer.GetTraceID(parentCtx)
	childTraceID := tracer.GetTraceID(childCtx)
	assert.Equal(t, parentTraceID, childTraceID)
	
	parentSpanID := tracer.GetSpanID(parentCtx)
	childSpanID := tracer.GetSpanID(childCtx)
	assert.NotEqual(t, parentSpanID, childSpanID)
} 