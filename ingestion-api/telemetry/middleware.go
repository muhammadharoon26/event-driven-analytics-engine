package telemetry

import (
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = "ingestion-api"

// TracingMiddleware creates a span for every incoming HTTP request and records
// standard HTTP semantic-convention attributes (method, route, status code).
// It also extracts any incoming W3C trace-context headers so that distributed
// traces are properly stitched together.
func TracingMiddleware() gin.HandlerFunc {
	tracer := otel.Tracer(tracerName)
	propagator := otel.GetTextMapPropagator()

	return func(c *gin.Context) {
		// Extract trace context from incoming request headers.
		ctx := propagator.Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))

		// Build a descriptive span name: "HTTP <METHOD> <path>".
		spanName := fmt.Sprintf("HTTP %s %s", c.Request.Method, c.FullPath())

		ctx, span := tracer.Start(ctx, spanName,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				attribute.String("http.method", c.Request.Method),
				attribute.String("http.route", c.FullPath()),
				attribute.String("http.url", c.Request.URL.String()),
				attribute.String("http.user_agent", c.Request.UserAgent()),
			),
		)
		defer span.End()

		// Replace the request context so downstream handlers can access the span.
		c.Request = c.Request.WithContext(ctx)

		// Inject trace context into response headers for downstream propagation.
		propagator.Inject(ctx, propagation.HeaderCarrier(c.Writer.Header()))

		// Process the rest of the middleware chain / handler.
		c.Next()

		// Record the final status code after the handler has executed.
		statusCode := c.Writer.Status()
		span.SetAttributes(attribute.Int("http.status_code", statusCode))

		// Mark the span as error when the server returns a 5xx.
		if statusCode >= 500 {
			span.SetAttributes(attribute.Bool("error", true))
		}
	}
}

// LatencyMiddleware measures wall-clock response time for every request.
// It writes the duration as an X-Response-Time header (human-readable) and
// records it as a span attribute so it shows up in the trace backend.
func LatencyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process the request.
		c.Next()

		elapsed := time.Since(start)

		// Expose the latency as a response header for easy verification.
		c.Writer.Header().Set("X-Response-Time", elapsed.String())

		// Attach latency to the active span (if one exists).
		span := trace.SpanFromContext(c.Request.Context())
		if span.IsRecording() {
			span.SetAttributes(
				attribute.Float64("http.response_time_ms", float64(elapsed.Microseconds())/1000.0),
			)
		}

		log.Printf("[latency] %s %s → %d | %s",
			c.Request.Method, c.Request.URL.Path, c.Writer.Status(), elapsed)
	}
}
