package tracing

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

// TracingConfig represents tracing configuration
type TracingConfig struct {
	Enabled         bool    `toml:"enabled"`
	ServiceName     string  `toml:"service_name"`
	ServiceVersion  string  `toml:"service_version"`
	Environment     string  `toml:"environment"`
	JaegerEndpoint  string  `toml:"jaeger_endpoint"`
	SamplingRate    float64 `toml:"sampling_rate"`
}

// OpenTelemetryTracer implements distributed tracing using OpenTelemetry
type OpenTelemetryTracer struct {
	logger       *logrus.Logger
	config       TracingConfig
	tracer       trace.Tracer
	provider     *tracesdk.TracerProvider
	serviceName  string
}

// NewOpenTelemetryTracer creates a new OpenTelemetry tracer
func NewOpenTelemetryTracer(logger *logrus.Logger, config TracingConfig) (*OpenTelemetryTracer, error) {
	if !config.Enabled {
		logger.Info("Tracing is disabled")
		return &OpenTelemetryTracer{
			logger: logger,
			config: config,
		}, nil
	}

	// Create Jaeger exporter
	var exp tracesdk.SpanExporter
	var err error
	
	if config.JaegerEndpoint != "" {
		exp, err = jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(config.JaegerEndpoint)))
		if err != nil {
			logger.WithError(err).Error("Failed to create Jaeger exporter")
			return nil, fmt.Errorf("failed to create Jaeger exporter: %w", err)
		}
		logger.WithField("endpoint", config.JaegerEndpoint).Info("Jaeger exporter initialized")
	} else {
		logger.Info("No Jaeger endpoint configured, tracing will be collected but not exported")
	}

	// Create resource
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(config.ServiceName),
			semconv.ServiceVersionKey.String(config.ServiceVersion),
			semconv.DeploymentEnvironmentKey.String(config.Environment),
		),
	)
	if err != nil {
		logger.WithError(err).Error("Failed to create tracing resource")
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create trace provider
	var tp *tracesdk.TracerProvider
	if exp != nil {
		tp = tracesdk.NewTracerProvider(
			tracesdk.WithBatcher(exp),
			tracesdk.WithResource(res),
			tracesdk.WithSampler(tracesdk.TraceIDRatioBased(config.SamplingRate)),
		)
	} else {
		// Create provider without exporter for local development
		tp = tracesdk.NewTracerProvider(
			tracesdk.WithResource(res),
			tracesdk.WithSampler(tracesdk.TraceIDRatioBased(config.SamplingRate)),
		)
	}

	// Register as global tracer provider
	otel.SetTracerProvider(tp)

	// Set global propagator
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	tracer := tp.Tracer(config.ServiceName)

	logger.WithFields(logrus.Fields{
		"service_name":    config.ServiceName,
		"service_version": config.ServiceVersion,
		"environment":     config.Environment,
		"sampling_rate":   config.SamplingRate,
	}).Info("OpenTelemetry tracing initialized")

	return &OpenTelemetryTracer{
		logger:      logger,
		config:      config,
		tracer:      tracer,
		provider:    tp,
		serviceName: config.ServiceName,
	}, nil
}

// StartSpan starts a new span with the given operation name
func (ott *OpenTelemetryTracer) StartSpan(ctx context.Context, operationName string) (context.Context, func()) {
	if !ott.config.Enabled || ott.tracer == nil {
		// Return a no-op function if tracing is disabled
		return ctx, func() {}
	}

	spanCtx, span := ott.tracer.Start(ctx, operationName)
	
	// Add common attributes
	span.SetAttributes(
		attribute.String("service.name", ott.serviceName),
		attribute.String("operation.name", operationName),
	)

	return spanCtx, func() {
		span.End()
	}
}

// StartSpanWithAttributes starts a new span with custom attributes
func (ott *OpenTelemetryTracer) StartSpanWithAttributes(ctx context.Context, operationName string, attrs map[string]string) (context.Context, func()) {
	if !ott.config.Enabled || ott.tracer == nil {
		return ctx, func() {}
	}

	spanCtx, span := ott.tracer.Start(ctx, operationName)
	
	// Add service attributes
	span.SetAttributes(
		attribute.String("service.name", ott.serviceName),
		attribute.String("operation.name", operationName),
	)

	// Add custom attributes
	for key, value := range attrs {
		span.SetAttributes(attribute.String(key, value))
	}

	return spanCtx, func() {
		span.End()
	}
}

// AddSpanEvent adds an event to the current span
func (ott *OpenTelemetryTracer) AddSpanEvent(ctx context.Context, name string, attributes map[string]string) {
	if !ott.config.Enabled {
		return
	}

	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	// Convert string attributes to attribute.KeyValue
	attrs := make([]attribute.KeyValue, 0, len(attributes))
	for key, value := range attributes {
		attrs = append(attrs, attribute.String(key, value))
	}

	span.AddEvent(name, trace.WithAttributes(attrs...))
}

// AddSpanError adds an error to the current span
func (ott *OpenTelemetryTracer) AddSpanError(ctx context.Context, err error) {
	if !ott.config.Enabled {
		return
	}

	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
}

// SetSpanStatus sets the status of the current span
func (ott *OpenTelemetryTracer) SetSpanStatus(ctx context.Context, code codes.Code, description string) {
	if !ott.config.Enabled {
		return
	}

	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	span.SetStatus(code, description)
}

// GetTraceID returns the trace ID of the current span
func (ott *OpenTelemetryTracer) GetTraceID(ctx context.Context) string {
	if !ott.config.Enabled {
		return ""
	}

	span := trace.SpanFromContext(ctx)
	if !span.SpanContext().IsValid() {
		return ""
	}

	return span.SpanContext().TraceID().String()
}

// GetSpanID returns the span ID of the current span
func (ott *OpenTelemetryTracer) GetSpanID(ctx context.Context) string {
	if !ott.config.Enabled {
		return ""
	}

	span := trace.SpanFromContext(ctx)
	if !span.SpanContext().IsValid() {
		return ""
	}

	return span.SpanContext().SpanID().String()
}

// Close shuts down the tracer provider
func (ott *OpenTelemetryTracer) Close() error {
	if ott.provider != nil {
		return ott.provider.Shutdown(context.Background())
	}
	return nil
}

// TraceExportOperation traces an export operation
func (ott *OpenTelemetryTracer) TraceExportOperation(ctx context.Context, exportType string, fn func(ctx context.Context) error) error {
	spanCtx, endSpan := ott.StartSpanWithAttributes(ctx, "export_operation", map[string]string{
		"export.type": exportType,
		"component":   "exporter",
	})
	defer endSpan()

	ott.AddSpanEvent(spanCtx, "export.started", map[string]string{
		"export.type": exportType,
	})

	err := fn(spanCtx)
	
	if err != nil {
		ott.AddSpanError(spanCtx, err)
		ott.AddSpanEvent(spanCtx, "export.failed", map[string]string{
			"export.type": exportType,
			"error":       err.Error(),
		})
		return err
	}

	ott.AddSpanEvent(spanCtx, "export.completed", map[string]string{
		"export.type": exportType,
	})
	ott.SetSpanStatus(spanCtx, codes.Ok, "Export completed successfully")

	return nil
}

// TraceAPICall traces an API call
func (ott *OpenTelemetryTracer) TraceAPICall(ctx context.Context, service, endpoint, method string, fn func(ctx context.Context) error) error {
	spanCtx, endSpan := ott.StartSpanWithAttributes(ctx, "api_call", map[string]string{
		"http.method":  method,
		"http.url":     endpoint,
		"service.name": service,
		"component":    "http_client",
	})
	defer endSpan()

	ott.AddSpanEvent(spanCtx, "api.request.started", map[string]string{
		"http.method": method,
		"http.url":    endpoint,
		"service":     service,
	})

	err := fn(spanCtx)
	
	if err != nil {
		ott.AddSpanError(spanCtx, err)
		ott.AddSpanEvent(spanCtx, "api.request.failed", map[string]string{
			"http.method": method,
			"http.url":    endpoint,
			"service":     service,
			"error":       err.Error(),
		})
		return err
	}

	ott.AddSpanEvent(spanCtx, "api.request.completed", map[string]string{
		"http.method": method,
		"http.url":    endpoint,
		"service":     service,
	})
	ott.SetSpanStatus(spanCtx, codes.Ok, "API call completed successfully")

	return nil
}

// TraceFileOperation traces a file operation
func (ott *OpenTelemetryTracer) TraceFileOperation(ctx context.Context, operation, filePath string, fn func(ctx context.Context) error) error {
	spanCtx, endSpan := ott.StartSpanWithAttributes(ctx, "file_operation", map[string]string{
		"file.operation": operation,
		"file.path":      filePath,
		"component":      "filesystem",
	})
	defer endSpan()

	ott.AddSpanEvent(spanCtx, "file.operation.started", map[string]string{
		"file.operation": operation,
		"file.path":      filePath,
	})

	err := fn(spanCtx)
	
	if err != nil {
		ott.AddSpanError(spanCtx, err)
		ott.AddSpanEvent(spanCtx, "file.operation.failed", map[string]string{
			"file.operation": operation,
			"file.path":      filePath,
			"error":          err.Error(),
		})
		return err
	}

	ott.AddSpanEvent(spanCtx, "file.operation.completed", map[string]string{
		"file.operation": operation,
		"file.path":      filePath,
	})
	ott.SetSpanStatus(spanCtx, codes.Ok, "File operation completed successfully")

	return nil
} 