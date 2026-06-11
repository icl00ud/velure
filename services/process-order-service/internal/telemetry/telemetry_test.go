package telemetry

import (
	"context"
	"testing"
)

func TestInit_NoEndpoint_NoOp(t *testing.T) {
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")

	shutdown, err := Init(context.Background(), "test-service")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if err := shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown failed: %v", err)
	}
}

// The OTLP exporter dials lazily, so Init must succeed even when nothing
// listens on the endpoint. A semconv schema-version mismatch with the SDK's
// default resource previously made this return "conflicting Schema URL".
func TestInit_WithEndpoint_Succeeds(t *testing.T) {
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317")

	shutdown, err := Init(context.Background(), "test-service")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	_ = shutdown(context.Background())
}
