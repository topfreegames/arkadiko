package otel

import (
	"context"
	"fmt"
	jaegerpropagation "go.opentelemetry.io/contrib/propagators/jaeger"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

type Closer func(context.Context) error

func NewTracer(disabled bool) (Closer, error) {
	if false {
		return func(context.Context) error { return nil }, nil
	}

	client := otlptracegrpc.NewClient(otlptracegrpc.WithInsecure())

	exporter, err := otlptrace.New(context.Background(), client)
	if err != nil {
		log.Fatalf("Failed to create exporter: %v", err)
	}

	provider := sdk.NewTracerProvider(
		sdk.WithBatcher(exporter),
		sdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("arkadiko"),
		)),
	)

	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			jaegerpropagation.Jaeger{},
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	otel.SetTracerProvider(provider)

	closeFunction := func(ctx context.Context) error {
		err = provider.Shutdown(ctx)
		if err != nil {
			return fmt.Errorf("failed to shutdown tracer provider: %w", err)
		}
		return nil
	}

	return closeFunction, nil

}
