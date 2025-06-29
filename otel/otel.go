package otel

import (
	"context"
	"fmt"

	jaegerpropagation "go.opentelemetry.io/contrib/propagators/jaeger"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

type Closer func(context.Context) error

type OtelImpl struct {
	serviceName string
	host        string
	port        string
	CloserFunc  Closer
}

const (
	DefaultOtelSamplingProbability = 0.1
)

func NewOtelImpl(ctx context.Context, serviceName, host, port string, samplingProbability float64) (*OtelImpl, error) {
	if samplingProbability == 0 {
		samplingProbability = DefaultOtelSamplingProbability
	}
	closer, err := NewTracer(ctx, serviceName, host, port, samplingProbability)
	if err != nil {
		return nil, err
	}
	return &OtelImpl{serviceName: serviceName, host: host, port: port, CloserFunc: closer}, nil
}

func NewTracer(ctx context.Context, serviceName, host, port string, samplingProbability float64) (Closer, error) {
	client := otlptracegrpc.NewClient(
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(fmt.Sprintf("%s:%s", host, port)),
	)

	exporter, err := otlptrace.New(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("failed to create exporter: %w", err)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create a sampler based on the sampling probability
	sampler := sdktrace.TraceIDRatioBased(samplingProbability)

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
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

func (o *OtelImpl) Tracer(name string) trace.Tracer {
	return otel.GetTracerProvider().Tracer(name)
}

func (o *OtelImpl) StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return o.Tracer(o.serviceName).Start(ctx, name, opts...)
}

func (o *OtelImpl) WithSpan(ctx context.Context, name string, fn func(context.Context) error) error {
	ctx, span := o.StartSpan(ctx, name)
	defer span.End()
	return fn(ctx)
}

func (o *OtelImpl) WithSpanAttributes(ctx context.Context, name string, attrs []attribute.KeyValue, fn func(context.Context) error) error {
	ctx, span := o.StartSpan(ctx, name, trace.WithAttributes(attrs...))
	defer span.End()
	return fn(ctx)
}

func GetSpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

// AddSpanAttributes adds attributes to the current span if available
func AddSpanAttributes(ctx context.Context, attrs ...attribute.KeyValue) {
	span := GetSpanFromContext(ctx)
	if span.IsRecording() {
		span.SetAttributes(attrs...)
	}
}

// AddSpanEvent adds an event to the current span if available
func AddSpanEvent(ctx context.Context, name string, attrs ...attribute.KeyValue) {
	span := GetSpanFromContext(ctx)
	if span.IsRecording() {
		span.AddEvent(name, trace.WithAttributes(attrs...))
	}
}

// SetSpanStatus sets the status of the current span if available
func SetSpanStatus(ctx context.Context, code codes.Code, description string) {
	span := GetSpanFromContext(ctx)
	if span.IsRecording() {
		span.SetStatus(code, description)
	}
}
