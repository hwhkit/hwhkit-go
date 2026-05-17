package observability

import (
	"context"
	"fmt"

	"github.com/hwhkit/hwhkit-go/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

type TracingShutdown func(ctx context.Context) error

func InitTracing(ctx context.Context, cfg config.OTelConfig) (*sdktrace.TracerProvider, TracingShutdown, error) {
	if !cfg.Enabled {
		return nil, func(context.Context) error { return nil }, nil
	}

	opts := []otlptracegrpc.Option{otlptracegrpc.WithEndpoint(cfg.Endpoint)}
	if cfg.Insecure {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}
	if len(cfg.Headers) > 0 {
		opts = append(opts, otlptracegrpc.WithHeaders(cfg.Headers))
	}

	exp, err := otlptrace.New(ctx, otlptracegrpc.NewClient(opts...))
	if err != nil {
		return nil, nil, fmt.Errorf("otlp: %w", err)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
		),
		resource.WithProcess(),
		resource.WithHost(),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("otel resource: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp, tp.Shutdown, nil
}
