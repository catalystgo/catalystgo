package tracer

import (
	"context"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/zipkin"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// newZipkinTracer creates a new Zipkin SpanExporter
func newZipkinTracer(_ context.Context, url string) (sdktrace.SpanExporter, error) {
	exp, err := zipkin.New(url)
	if err != nil {
		return nil, err
	}
	return exp, nil
}

// newJaegerTracer creates a new Jaeger SpanExporter
func newJaegerTracer(ctx context.Context, url string) (sdktrace.SpanExporter, error) {
	exp, err := otlptracegrpc.New(ctx, otlptracegrpc.WithEndpoint(url))
	if err != nil {
		return nil, err
	}
	return exp, nil
}
