// Package telemetry wires OpenTelemetry tracing. Tracing is enabled when
// OTEL_EXPORTER_OTLP_ENDPOINT is set (host:port of an OTLP gRPC collector,
// e.g. Jaeger); otherwise the global no-op tracer stays in place and the
// propagation helpers still work, so instrumented code needs no branching.
package telemetry

import (
	"context"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// Init configures the global tracer provider and W3C propagators.
// The returned shutdown func flushes pending spans; it is never nil.
func Init(ctx context.Context, serviceName string) (func(context.Context) error, error) {
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		return func(context.Context) error { return nil }, nil
	}

	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return func(context.Context) error { return nil }, err
	}

	// Schemaless avoids "conflicting Schema URL" errors when the semconv
	// package version diverges from the SDK's default resource schema.
	res := sdkresource.NewSchemaless(semconv.ServiceName(serviceName))

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)

	return tp.Shutdown, nil
}

// InjectMap serializes the trace context from ctx into a string map, suitable
// for AMQP headers or an outbox column.
func InjectMap(ctx context.Context) map[string]string {
	carrier := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, carrier)
	return carrier
}

// ExtractMap restores a trace context previously captured with InjectMap.
func ExtractMap(ctx context.Context, m map[string]string) context.Context {
	if len(m) == 0 {
		return ctx
	}
	return otel.GetTextMapPropagator().Extract(ctx, propagation.MapCarrier(m))
}

// Traceparent returns the W3C traceparent header value for ctx, or "" when
// there is no active span.
func Traceparent(ctx context.Context) string {
	return InjectMap(ctx)["traceparent"]
}

// WithTraceparent restores a context from a stored traceparent value.
func WithTraceparent(ctx context.Context, traceparent string) context.Context {
	if traceparent == "" {
		return ctx
	}
	return ExtractMap(ctx, map[string]string{"traceparent": traceparent})
}
