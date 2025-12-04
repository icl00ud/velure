package handlers_test

import (
	"net/http/httptest"
	"testing"

	"product-service/internal/handler"

	"github.com/gofiber/fiber/v2"
)

func TestHealthCheck(t *testing.T) {
	app := fiber.New()
	healthHandler := handlers.NewHealthHandler()

	app.Get("/health", healthHandler.Check)

	req := httptest.NewRequest("GET", "/health", nil)
	resp, err := app.Test(req)

	if err != nil {
		t.Fatalf("Test failed: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}
}
