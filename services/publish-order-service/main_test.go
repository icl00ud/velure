package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"syscall"
	"testing"
	"time"

	"github.com/icl00ud/velure-shared/logger"

	"github.com/icl00ud/publish-order-service/internal/config"
	"github.com/icl00ud/publish-order-service/internal/consumer"
	"github.com/icl00ud/publish-order-service/internal/model"
	"github.com/icl00ud/publish-order-service/internal/publisher"
	"github.com/icl00ud/publish-order-service/internal/repository"
)

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
	// Touch defaultDeps to keep coverage meaningful without invoking real connections.
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
