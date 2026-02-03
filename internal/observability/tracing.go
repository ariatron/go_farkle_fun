package observability

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

// Tracer is the global tracer instance
var Tracer trace.Tracer

// tracerProvider is the global tracer provider
var tracerProvider *sdktrace.TracerProvider

// InitTracing initializes OpenTelemetry tracing with OTLP exporter
// endpoint should be in the format "localhost:4318" for HTTP
func InitTracing(serviceName, endpoint string) error {
	ctx := context.Background()

	// Create OTLP HTTP exporter
	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithInsecure(), // For local development
	)
	if err != nil {
		return fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// Create resource with service information
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion("1.0.0"),
			attribute.String("environment", "development"),
		),
	)
	if err != nil {
		return fmt.Errorf("failed to create resource: %w", err)
	}

	// Create tracer provider
	tracerProvider = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	// Set global tracer provider
	otel.SetTracerProvider(tracerProvider)

	// Create tracer
	Tracer = tracerProvider.Tracer("farkle-app")

	Logger.Info("Tracing initialized",
		"service", serviceName,
		"endpoint", endpoint,
		"exporter", "otlp-http",
	)

	return nil
}

// ShutdownTracing gracefully shuts down the tracer provider
func ShutdownTracing(ctx context.Context) error {
	if tracerProvider == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := tracerProvider.Shutdown(ctx); err != nil {
		Logger.Error("Failed to shutdown tracer provider", "error", err)
		return err
	}

	Logger.Info("Tracing shutdown complete")
	return nil
}

// StartSpan is a helper to start a span with common attributes
func StartSpan(ctx context.Context, name string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	return Tracer.Start(ctx, name, trace.WithAttributes(attrs...))
}

// AddSpanEvent adds an event to the current span
func AddSpanEvent(ctx context.Context, name string, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.AddEvent(name, trace.WithAttributes(attrs...))
	}
}

// AddSpanError adds an error to the current span
func AddSpanError(ctx context.Context, err error) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.RecordError(err)
	}
}

// GetTraceID extracts the trace ID from the context
func GetTraceID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		return span.SpanContext().TraceID().String()
	}
	return ""
}
