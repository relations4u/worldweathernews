package observability

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// TracingConfig steuert den OTLP-Exporter und die Resource-Attribute.
// Endpoint ist host:port ohne Schema (gRPC, OTLP-Default 4317).
type TracingConfig struct {
	Enabled     bool
	Endpoint    string
	ServiceName string
	Environment string
	Version     string
}

// ShutdownFunc trägt die Shutdown-Logik des Tracer-Providers — vom Caller
// in der Shutdown-Sequenz aufzurufen, damit batched Spans noch flushen.
type ShutdownFunc func(context.Context) error

// InitTracing setzt den globalen Tracer-Provider. Bei Enabled=false
// liefert es eine no-op ShutdownFunc und ändert keinen globalen State,
// damit Tests/CLI ohne Tempo laufen.
//
// AlwaysSample ist für Dev OK. TODO(prod): ParentBased(TraceIDRatioBased)
// sobald echte Production läuft.
func InitTracing(ctx context.Context, cfg TracingConfig) (ShutdownFunc, error) {
	if !cfg.Enabled {
		return func(context.Context) error { return nil }, nil
	}

	exp, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(cfg.Endpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("create OTLP exporter: %w", err)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.Version),
			semconv.DeploymentEnvironment(cfg.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("create resource: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return func(shutdownCtx context.Context) error {
		shutdownCtx, cancel := context.WithTimeout(shutdownCtx, 5*time.Second)
		defer cancel()
		return tp.Shutdown(shutdownCtx)
	}, nil
}
