package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/icl00ud/publish-order-service/internal/handler"
	"github.com/icl00ud/publish-order-service/internal/middleware"
	"github.com/icl00ud/velure-shared/logger"

	"github.com/icl00ud/publish-order-service/internal/config"
	"github.com/icl00ud/publish-order-service/internal/consumer"
	"github.com/icl00ud/publish-order-service/internal/model"
	"github.com/icl00ud/publish-order-service/internal/publisher"
	"github.com/icl00ud/publish-order-service/internal/repository"
)

func TestRegisterRoutes_CanonicalOnly(t *testing.T) {
	const jwtSecret = "test-secret"

	svc := &routingStubService{}
	oh := handler.NewOrderHandler(svc, &stubPublisher{})
	sse := handler.NewSSEHandler(svc)

	mux := http.NewServeMux()
	registerRoutes(mux, oh, sse, middleware.Auth(jwtSecret), middleware.SSEAuth(jwtSecret))

	authToken := issueTestJWT(t, jwtSecret, "user-1")

	tests := []struct {
		name           string
		method         string
		target         string
		body           string
		authHeader     string
		wantStatusCode int
		wantLastID     string
	}{
		{name: "legacy update alias removed", method: http.MethodPost, target: "/update-order-status", body: `{"order_id":"legacy-1","status":"PROCESSING"}`, wantStatusCode: http.StatusNotFound},
		{name: "legacy create alias removed", method: http.MethodPost, target: "/create-order", body: `[{"product_id":"p1","quantity":1}]`, wantStatusCode: http.StatusNotFound},
		{name: "root create route removed", method: http.MethodPost, target: "/orders", body: `[{"product_id":"p1","quantity":1}]`, wantStatusCode: http.StatusNotFound},
		{name: "canonical create requires auth", method: http.MethodPost, target: "/api/orders", body: `[{"product_id":"p1","quantity":1}]`, wantStatusCode: http.StatusUnauthorized},
		{name: "root list route removed", method: http.MethodGet, target: "/orders", wantStatusCode: http.StatusNotFound},
		{name: "canonical list orders", method: http.MethodGet, target: "/api/orders", wantStatusCode: http.StatusOK},
		{name: "root me orders route removed", method: http.MethodGet, target: "/me/orders", wantStatusCode: http.StatusNotFound},
		{name: "canonical me orders requires auth", method: http.MethodGet, target: "/api/me/orders", wantStatusCode: http.StatusUnauthorized},
		{name: "root me order by id route removed", method: http.MethodGet, target: "/me/orders/order-123", authHeader: "Bearer " + authToken, wantStatusCode: http.StatusNotFound},
		{name: "canonical me order by id injects query", method: http.MethodGet, target: "/api/me/orders/order-456", authHeader: "Bearer " + authToken, wantStatusCode: http.StatusOK, wantLastID: "order-456"},
		{name: "root events route removed", method: http.MethodGet, target: "/me/orders/order-789/events?token=" + authToken, wantStatusCode: http.StatusNotFound},
		{name: "canonical events injects query", method: http.MethodGet, target: "/api/me/orders/order-890/events?token=" + authToken, wantStatusCode: http.StatusOK, wantLastID: "order-890"},
		{name: "root patch update route removed", method: http.MethodPatch, target: "/orders/order-123/status", body: `{"order_id":"order-123","status":"COMPLETED"}`, wantStatusCode: http.StatusNotFound},
		{name: "canonical patch update", method: http.MethodPatch, target: "/api/orders/order-456/status", body: `{"order_id":"order-456","status":"FAILED"}`, wantStatusCode: http.StatusOK},
		{name: "legacy api order namespace removed", method: http.MethodPost, target: "/api/order/create-order", body: `[{"product_id":"p1","quantity":1}]`, wantStatusCode: http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			req := httptest.NewRequest(tt.method, tt.target, strings.NewReader(tt.body)).WithContext(ctx)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			if tt.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}

			if strings.Contains(tt.target, "/events") {
				cancel()
			}

			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatusCode {
				t.Fatalf("unexpected status: got %d want %d", rr.Code, tt.wantStatusCode)
			}
			if tt.wantLastID != "" && svc.lastGetOrderByID != tt.wantLastID {
				t.Fatalf("expected id %q to be passed to handler, got %q", tt.wantLastID, svc.lastGetOrderByID)
			}
		})
	}
}

func issueTestJWT(t *testing.T, secret, subject string) string {
	t.Helper()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{Subject: subject})
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to sign jwt: %v", err)
	}

	return signed
}

type routingStubService struct {
	lastGetOrderByID string
}

func (s *routingStubService) Create(_ context.Context, userID string, items []model.CartItem) (model.Order, error) {
	return model.Order{ID: "created-1", UserID: userID, Items: items, Status: model.StatusCreated}, nil
}

func (s *routingStubService) UpdateStatus(_ context.Context, id, status string) (model.Order, error) {
	return model.Order{ID: id, Status: status}, nil
}

func (s *routingStubService) GetOrdersByPage(_ context.Context, page, pageSize int) (*model.PaginatedOrdersResponse, error) {
	return &model.PaginatedOrdersResponse{Orders: []model.Order{}, Page: page, PageSize: pageSize, TotalPages: 1}, nil
}

func (s *routingStubService) GetOrdersByUserID(_ context.Context, userID string, page, pageSize int) (*model.PaginatedOrdersResponse, error) {
	return &model.PaginatedOrdersResponse{Orders: []model.Order{}, Page: page, PageSize: pageSize, TotalPages: 1}, nil
}

func (s *routingStubService) GetOrderByID(_ context.Context, userID, orderID string) (model.Order, error) {
	s.lastGetOrderByID = orderID
	return model.Order{ID: orderID, UserID: userID, Status: model.StatusCreated}, nil
}

func TestRunWithInjectedDependencies(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	cfg := config.Config{
		Port:        "8080",
		PostgresURL: "postgres://example",
		RabbitURL:   "amqp://example",
		Exchange:    "orders",
		Queue:       "queue",
		Workers:     1,
		JWTSecret:   "secret",
	}

	fakeRepo := &stubRepository{db: &sql.DB{}}
	fakePublisher := &stubPublisher{}
	fakeConsumer := &stubConsumer{}
	fakeServer := newStubServer()

	deps := appDeps{
		loadConfig: func() (config.Config, error) {
			return cfg, nil
		},
		newRepo: func(dsn string) (repository.OrderRepository, error) {
			if dsn != cfg.PostgresURL {
				return nil, fmt.Errorf("unexpected dsn: %s", dsn)
			}
			return fakeRepo, nil
		},
		runMigrations: func(db *sql.DB, path string) error {
			fakeRepo.migrationsRan = true
			fakeRepo.migrationsPath = path
			return nil
		},
		newPublisher: func(url, exchange string, logger *logger.Logger) (publisher.Publisher, error) {
			fakePublisher.url = url
			fakePublisher.exchange = exchange
			return fakePublisher, nil
		},
		newConsumer: func(url, exchange, queue string, handler consumer.EventHandler, workers int, logger *logger.Logger) (consumer.Consumer, error) {
			fakeConsumer.url = url
			fakeConsumer.exchange = exchange
			fakeConsumer.queue = queue
			fakeConsumer.workers = workers
			fakeConsumer.handler = handler
			return fakeConsumer, nil
		},
		newLogger: func() *logger.Logger {
			return logger.NewNop()
		},
		newHTTPServer: func(cfg config.Config, handler http.Handler) server {
			fakeServer.handler = handler
			return fakeServer
		},
	}

	if err := run(ctx, deps); err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	if !fakeRepo.migrationsRan {
		t.Fatal("expected migrations to run")
	}
	if fakeRepo.migrationsPath != "./internal/migrations" {
		t.Fatalf("unexpected migrations path: %s", fakeRepo.migrationsPath)
	}
	if !fakeServer.started || !fakeServer.shutDown {
		t.Fatalf("server lifecycle not executed, started=%t shutdown=%t", fakeServer.started, fakeServer.shutDown)
	}
	if !fakeConsumer.started || !fakeConsumer.closed {
		t.Fatalf("consumer lifecycle not executed, started=%t closed=%t", fakeConsumer.started, fakeConsumer.closed)
	}
	if !fakePublisher.closed {
		t.Fatal("publisher was not closed")
	}
	if fakeConsumer.handler == nil {
		t.Fatal("event handler was not wired into consumer")
	}
}

func TestMainUsesInjectedFactoryAndHandlesSignal(t *testing.T) {
	if d := defaultDeps(); d.newRepo == nil {
		t.Fatal("default deps not initialized")
	}

	cfg := config.Config{
		Port:        "8081",
		PostgresURL: "postgres://testing",
		RabbitURL:   "amqp://testing",
		Exchange:    "orders",
		Queue:       "queue",
		Workers:     1,
		JWTSecret:   "secret",
	}

	fakeRepo := &stubRepository{db: &sql.DB{}}
	fakePublisher := &stubPublisher{}
	fakeConsumer := &stubConsumer{}
	fakeServer := newStubServer()

	originalFactory := depsFactory
	defer func() { depsFactory = originalFactory }()

	depsFactory = func() appDeps {
		return appDeps{
			loadConfig: func() (config.Config, error) { return cfg, nil },
			newRepo: func(string) (repository.OrderRepository, error) {
				return fakeRepo, nil
			},
			runMigrations: func(*sql.DB, string) error { return nil },
			newPublisher: func(string, string, *logger.Logger) (publisher.Publisher, error) {
				return fakePublisher, nil
			},
			newConsumer: func(string, string, string, consumer.EventHandler, int, *logger.Logger) (consumer.Consumer, error) {
				return fakeConsumer, nil
			},
			newLogger: func() *logger.Logger { return logger.NewNop() },
			newHTTPServer: func(config.Config, http.Handler) server {
				return fakeServer
			},
		}
	}

	go func() {
		time.Sleep(20 * time.Millisecond)
		_ = syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	}()

	main()

	if !fakeServer.started || !fakeServer.shutDown {
		t.Fatalf("server should start and shut down, started=%t shutdown=%t", fakeServer.started, fakeServer.shutDown)
	}
	if !fakeConsumer.started || !fakeConsumer.closed {
		t.Fatalf("consumer should start and close, started=%t closed=%t", fakeConsumer.started, fakeConsumer.closed)
	}
	if !fakePublisher.closed {
		t.Fatal("publisher should be closed at shutdown")
	}
}

func TestRunReturnsErrorWhenRepositoryCreationFails(t *testing.T) {
	deps := appDeps{
		loadConfig: func() (config.Config, error) {
			return config.Config{
				Port:        "8080",
				PostgresURL: "postgres://broken",
				RabbitURL:   "amqp://example",
				Exchange:    "orders",
				Queue:       "queue",
				Workers:     1,
				JWTSecret:   "secret",
			}, nil
		},
		newLogger: func() *logger.Logger { return logger.NewNop() },
		newRepo: func(string) (repository.OrderRepository, error) {
			return nil, errors.New("boom")
		},
		runMigrations: func(*sql.DB, string) error { return nil },
		newPublisher:  func(string, string, *logger.Logger) (publisher.Publisher, error) { return &stubPublisher{}, nil },
		newConsumer: func(string, string, string, consumer.EventHandler, int, *logger.Logger) (consumer.Consumer, error) {
			return &stubConsumer{}, nil
		},
		newHTTPServer: func(config.Config, http.Handler) server { return newStubServer() },
	}

	err := run(context.Background(), deps)
	if err == nil || err.Error() != "repository init failed: boom" {
		t.Fatalf("expected repository init failure, got %v", err)
	}
}

type stubRepository struct {
	db             *sql.DB
	migrationsRan  bool
	migrationsPath string
}

func (s *stubRepository) Save(context.Context, model.Order) error { return nil }

func (s *stubRepository) Find(context.Context, string) (model.Order, error) {
	return model.Order{}, nil
}

func (s *stubRepository) FindByUserID(context.Context, string, string) (model.Order, error) {
	return model.Order{}, nil
}

func (s *stubRepository) GetOrdersByPage(context.Context, int, int) (*model.PaginatedOrdersResponse, error) {
	return &model.PaginatedOrdersResponse{}, nil
}

func (s *stubRepository) GetOrdersByUserID(context.Context, string, int, int) (*model.PaginatedOrdersResponse, error) {
	return &model.PaginatedOrdersResponse{}, nil
}

func (s *stubRepository) GetOrdersCount(context.Context) (int64, error) { return 0, nil }

func (s *stubRepository) GetOrdersCountByUserID(context.Context, string) (int64, error) {
	return 0, nil
}

func (s *stubRepository) DB() *sql.DB { return s.db }

type stubPublisher struct {
	url      string
	exchange string
	closed   bool
}

func (s *stubPublisher) Publish(model.Event) error { return nil }

func (s *stubPublisher) Close() error {
	s.closed = true
	return nil
}

type stubConsumer struct {
	url      string
	exchange string
	queue    string
	workers  int
	started  bool
	closed   bool
	handler  consumer.EventHandler
}

func (s *stubConsumer) Start(ctx context.Context) error {
	s.started = true
	<-ctx.Done()
	return nil
}

func (s *stubConsumer) Close() error {
	s.closed = true
	return nil
}

type stubServer struct {
	handler     http.Handler
	shutdownCh  chan struct{}
	started     bool
	shutDown    bool
	serveResult error
}

func newStubServer() *stubServer {
	return &stubServer{
		shutdownCh: make(chan struct{}),
	}
}

func (s *stubServer) ListenAndServe() error {
	s.started = true
	<-s.shutdownCh
	if s.serveResult != nil {
		return s.serveResult
	}
	return http.ErrServerClosed
}

func (s *stubServer) Shutdown(context.Context) error {
	if s.shutDown {
		return nil
	}
	s.shutDown = true
	close(s.shutdownCh)
	return nil
}
