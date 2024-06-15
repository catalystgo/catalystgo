package tracer

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

var (
	errUnknownTraceProvider = func(tp string) error { return fmt.Errorf("unknown trace provider: %s", tp) }
)

type tracerProvider string

const (
	ZIPKIN = "zipkin"
	JAEGER = "jaeger"
)

// Init create a new tracing provider
func Init(ctx context.Context, provider tracerProvider, url string, opts ...sdktrace.TracerProviderOption) (func(ctx context.Context) error, error) {
	exp, err := getExporter(ctx, provider, url)
	if err != nil {
		return nil, err
	}

	opts = append(opts, sdktrace.WithBatcher(exp))
	tp := sdktrace.NewTracerProvider(opts...)

	otel.SetTracerProvider(tp)

	return tp.Shutdown, nil
}

func getExporter(ctx context.Context, provider tracerProvider, url string) (sdktrace.SpanExporter, error) {
	switch provider {
	case ZIPKIN:
		return newZipkinTracer(ctx, url)
	case JAEGER:
		return newJaegerTracer(ctx, url)
	default:
		return nil, errUnknownTraceProvider(string(provider))
	}
}
