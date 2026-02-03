package observability

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// responseWriter wraps http.ResponseWriter to capture response data
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

// ObservabilityMiddleware provides comprehensive logging, metrics, and tracing
func ObservabilityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Start tracing span
		ctx, span := StartSpan(r.Context(), r.Method+" "+r.URL.Path,
			attribute.String("http.method", r.Method),
			attribute.String("http.url", r.URL.String()),
			attribute.String("http.scheme", r.URL.Scheme),
			attribute.String("http.host", r.Host),
			attribute.String("http.user_agent", r.UserAgent()),
		)
		defer span.End()

		// Create wrapped response writer
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     200, // default status code
		}

		// Execute the handler with context containing trace info
		next.ServeHTTP(wrapped, r.WithContext(ctx))

		// Calculate duration
		duration := time.Since(start)
		durationSecs := duration.Seconds()

		// Get trace ID for logging
		traceID := GetTraceID(ctx)

		// Record metrics
		AppMetrics.RecordHTTPRequest(
			r.Method,
			r.URL.Path,
			wrapped.statusCode,
			durationSecs,
			wrapped.size,
		)

		// Add span attributes
		span.SetAttributes(
			attribute.Int("http.status_code", wrapped.statusCode),
			attribute.Int("http.response_size", wrapped.size),
			attribute.Float64("http.duration_ms", duration.Seconds()*1000),
		)

		// Set span status based on HTTP status code
		if wrapped.statusCode >= 400 {
			span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", wrapped.statusCode))
		} else {
			span.SetStatus(codes.Ok, "")
		}

		// Log request
		Logger.Info("HTTP request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", wrapped.statusCode,
			"duration_ms", duration.Milliseconds(),
			"size_bytes", wrapped.size,
			"trace_id", traceID,
			"user_agent", r.UserAgent(),
			"remote_addr", r.RemoteAddr,
		)
	})
}

// RecoveryMiddleware recovers from panics and logs them
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Get trace ID
				traceID := GetTraceID(r.Context())

				// Log panic
				Logger.Error("Panic recovered",
					"error", err,
					"method", r.Method,
					"path", r.URL.Path,
					"trace_id", traceID,
				)

				// Add error to span
				span := trace.SpanFromContext(r.Context())
				if span.IsRecording() {
					span.RecordError(fmt.Errorf("panic: %v", err))
					span.SetStatus(codes.Error, "panic")
				}

				// Return 500 error
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// CORSMiddleware adds CORS headers
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// HealthCheckMiddleware bypasses observability for health checks
func HealthCheckMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip observability for health checks to reduce noise
		if r.URL.Path == "/health" || r.URL.Path == "/healthz" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
			return
		}

		next.ServeHTTP(w, r)
	})
}

// statusCodeToString converts status code to string for labels
func statusCodeToString(code int) string {
	return strconv.Itoa(code)
}
