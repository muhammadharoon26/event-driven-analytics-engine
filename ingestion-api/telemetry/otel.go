package telemetry

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

// InitTracer initializes an OTLP exporter, and configures the corresponding trace and
// metric providers. If the OTEL Collector is unreachable, it returns a no-op provider
// so the API can still start without tracing.
func InitTracer() (*sdktrace.TracerProvider, error) {
	// Use a short timeout so the API doesn't hang if the collector is down.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		endpoint = "localhost:4317" // Default Jaeger/Collector
	}

	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(endpoint),
	)
	if err != nil {
		log.Printf("[telemetry] WARN: could not create OTLP exporter (%v) — tracing disabled", err)
		return newNoopProvider(), nil
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName("ingestion-api"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	log.Printf("[telemetry] OpenTelemetry tracer initialized → exporting to %s", endpoint)
	return tp, nil
}

// newNoopProvider returns a TracerProvider that records spans but drops them.
// This lets the middleware run without errors even when the collector is offline.
func newNoopProvider() *sdktrace.TracerProvider {
	tp := sdktrace.NewTracerProvider() // no exporter = spans are dropped
	otel.SetTracerProvider(tp)
	return tp
}
