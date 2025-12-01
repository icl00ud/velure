package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"velure-auth-service/internal/config"
	"velure-auth-service/internal/handlers"
	"velure-auth-service/internal/mocks"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/icl00ud/velure-shared/logger"
	"go.uber.org/mock/gomock"
)

func TestSetupRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockAuthServiceInterface(ctrl)
	handler := handlers.NewAuthHandler(mockService)

	router := gin.New()
	setupRoutes(router, handler)

	// Verify routes exist
	routes := router.Routes()
	if len(routes) == 0 {
		t.Error("Expected routes to be configured")
	}

	expectedRoutes := map[string]string{
		"POST /authentication/register":               "POST",
		"POST /authentication/login":                  "POST",
		"POST /authentication/validateToken":          "POST",
		"GET /authentication/users":                   "GET",
		"GET /authentication/user/id/:id":             "GET",
		"GET /authentication/user/email/:email":       "GET",
		"DELETE /authentication/logout/:refreshToken": "DELETE",
	}

	foundRoutes := make(map[string]bool)
	for _, route := range routes {
		key := route.Method + " " + route.Path
		foundRoutes[key] = true
	}

	for expectedRoute := range expectedRoutes {
		if !foundRoutes[expectedRoute] {
			t.Errorf("Expected route '%s' not found", expectedRoute)
		}
	}
}

func TestSetupRouter_Development(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfg := &config.Config{
		Environment: "development",
	}

	mockService := mocks.NewMockAuthServiceInterface(ctrl)
	handler := handlers.NewAuthHandler(mockService)

	router := setupRouter(cfg, handler)

	if router == nil {
		t.Fatal("setupRouter() returned nil")
	}

	// Verify routes are set up
	routes := router.Routes()
	if len(routes) < 7 { // At least auth routes
		t.Errorf("Expected at least 7 routes, got %d", len(routes))
	}
}

func TestSetupRouter_Production(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfg := &config.Config{
		Environment: "production",
	}

	mockService := mocks.NewMockAuthServiceInterface(ctrl)
	handler := handlers.NewAuthHandler(mockService)

	router := setupRouter(cfg, handler)

	if router == nil {
		t.Fatal("setupRouter() returned nil")
	}

	// Verify routes are set up
	routes := router.Routes()
	if len(routes) < 7 {
		t.Errorf("Expected at least 7 routes, got %d", len(routes))
	}
}

func TestSetupRouter_HealthEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfg := &config.Config{
		Environment: "development",
	}

	mockService := mocks.NewMockAuthServiceInterface(ctrl)
	handler := handlers.NewAuthHandler(mockService)

	router := setupRouter(cfg, handler)

	// Test health endpoint
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	expectedBody := `{"status":"ok"}`
	if w.Body.String() != expectedBody {
		t.Errorf("Expected body %s, got %s", expectedBody, w.Body.String())
	}
}

func TestSetupRouter_MetricsEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfg := &config.Config{
		Environment: "development",
	}

	mockService := mocks.NewMockAuthServiceInterface(ctrl)
	handler := handlers.NewAuthHandler(mockService)

	router := setupRouter(cfg, handler)

	// Test metrics endpoint exists
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Metrics endpoint should return 200
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Should contain prometheus metrics
	body := w.Body.String()
	if len(body) == 0 {
		t.Error("Expected metrics output, got empty body")
	}
}

func TestSetupRouter_HasMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfg := &config.Config{
		Environment: "development",
	}

	mockService := mocks.NewMockAuthServiceInterface(ctrl)
	handler := handlers.NewAuthHandler(mockService)

	router := setupRouter(cfg, handler)

	// Make a request to test that middleware is applied
	req := httptest.NewRequest("OPTIONS", "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// CORS middleware should set headers
	corsHeader := w.Header().Get("Access-Control-Allow-Origin")
	if corsHeader == "" {
		t.Error("Expected CORS headers to be set by middleware")
	}
}

func TestRun_WithSQLiteAndRedis(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRedis := miniredis.RunT(t)

	t.Setenv("AUTH_SERVICE_SKIP_HTTP", "true")
	t.Setenv("POSTGRES_URL", "sqlite://file::memory:?cache=shared")
	t.Setenv("REDIS_HOST", mockRedis.Host())
	t.Setenv("REDIS_PORT", mockRedis.Port())

	if err := run(logger.NewNop()); err != nil {
		t.Fatalf("run(logger.NewNop()) returned unexpected error: %v", err)
	}
}

func TestMain_SkipHTTP(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRedis := miniredis.RunT(t)

	t.Setenv("AUTH_SERVICE_SKIP_HTTP", "true")
	t.Setenv("POSTGRES_URL", "sqlite://file::memory:?cache=shared")
	t.Setenv("REDIS_HOST", mockRedis.Host())
	t.Setenv("REDIS_PORT", mockRedis.Port())

	main()
}

func TestConnectRedis_Error(t *testing.T) {
	_, err := connectRedis(config.RedisConfig{
		Addr: "127.0.0.1:0",
	})
	if err == nil {
		t.Fatalf("expected error connecting to invalid redis address")
	}
}
