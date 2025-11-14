package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestPrometheusMiddleware_Success(t *testing.T) {
	app := fiber.New()

	// Add prometheus middleware
	app.Use(PrometheusMiddleware())

	// Add test route
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestPrometheusMiddleware_WithError(t *testing.T) {
	app := fiber.New()

	// Add prometheus middleware
	app.Use(PrometheusMiddleware())

	// Add test route that returns an error
	app.Get("/error", func(c *fiber.Ctx) error {
		return fiber.NewError(fiber.StatusBadRequest, "test error")
	})

	req := httptest.NewRequest("GET", "/error", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestPrometheusMiddleware_NotFound(t *testing.T) {
	app := fiber.New()

	// Add prometheus middleware
	app.Use(PrometheusMiddleware())

	req := httptest.NewRequest("GET", "/notfound", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 404, resp.StatusCode)
}

func TestPrometheusMiddleware_POST(t *testing.T) {
	app := fiber.New()

	// Add prometheus middleware
	app.Use(PrometheusMiddleware())

	// Add test route
	app.Post("/create", func(c *fiber.Ctx) error {
		return c.Status(201).SendString("Created")
	})

	req := httptest.NewRequest("POST", "/create", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode)
}

func TestPrometheusMiddleware_DELETE(t *testing.T) {
	app := fiber.New()

	// Add prometheus middleware
	app.Use(PrometheusMiddleware())

	// Add test route
	app.Delete("/delete/:id", func(c *fiber.Ctx) error {
		return c.SendStatus(204)
	})

	req := httptest.NewRequest("DELETE", "/delete/123", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 204, resp.StatusCode)
}

func TestPrometheusMiddleware_PUT(t *testing.T) {
	app := fiber.New()

	// Add prometheus middleware
	app.Use(PrometheusMiddleware())

	// Add test route
	app.Put("/update/:id", func(c *fiber.Ctx) error {
		return c.SendString("Updated")
	})

	req := httptest.NewRequest("PUT", "/update/123", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestPrometheusMiddleware_MultipleRequests(t *testing.T) {
	app := fiber.New()

	// Add prometheus middleware
	app.Use(PrometheusMiddleware())

	// Add test routes
	app.Get("/route1", func(c *fiber.Ctx) error {
		return c.SendString("Route 1")
	})
	app.Get("/route2", func(c *fiber.Ctx) error {
		return c.SendString("Route 2")
	})

	// Make multiple requests
	req1 := httptest.NewRequest("GET", "/route1", nil)
	resp1, err1 := app.Test(req1)

	assert.NoError(t, err1)
	assert.Equal(t, 200, resp1.StatusCode)

	req2 := httptest.NewRequest("GET", "/route2", nil)
	resp2, err2 := app.Test(req2)

	assert.NoError(t, err2)
	assert.Equal(t, 200, resp2.StatusCode)
}

func TestPrometheusMiddleware_InternalServerError(t *testing.T) {
	app := fiber.New()

	// Add prometheus middleware
	app.Use(PrometheusMiddleware())

	// Add test route that returns 500
	app.Get("/error500", func(c *fiber.Ctx) error {
		return fiber.NewError(fiber.StatusInternalServerError, "internal error")
	})

	req := httptest.NewRequest("GET", "/error500", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode)
}
