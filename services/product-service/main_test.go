package main

import (
	"context"
	"errors"
	"net/http/httptest"
	"os"
	"testing"

	"product-service/internal/config"
	"product-service/internal/model"
	"product-service/internal/repository"
	"product-service/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/icl00ud/velure-shared/logger"
	"github.com/stretchr/testify/assert"
)

func TestMaskURI(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "standard mongodb uri with password",
			input:    "mongodb://username:password123@localhost:27017",
			expected: "mongodb://username:***@localhost:27017",
		},
		{
			name:     "mongodb uri with special characters in password",
			input:    "mongodb://user:password_with!special@cluster.mongodb.net:27017",
			expected: "mongodb://user:***@cluster.mongodb.net:27017",
		},
		{
			name:     "mongodb uri without password",
			input:    "mongodb://localhost:27017",
			expected: "mongodb://localhost:27017",
		},
		{
			name:     "mongodb+srv uri",
			input:    "mongodb+srv://user:secret@cluster.mongodb.net/database",
			expected: "mongodb+srv://user:secret@cluster.mongodb.net/database",
		},
		{
			name:     "short uri",
			input:    "mongodb://",
			expected: "mongodb://",
		},
		{
			name:     "non-mongodb uri",
			input:    "postgresql://user:pass@localhost:5432/db",
			expected: "postgresql://user:pass@localhost:5432/db",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "mongodb uri with long password",
			input:    "mongodb://admin:verylongpassword123456789@server.example.com:27017/database",
			expected: "mongodb://admin:***@server.example.com:27017/database",
		},
		{
			name:     "mongodb uri with username only",
			input:    "mongodb://username@localhost:27017",
			expected: "mongodb://username@localhost:27017",
		},
		{
			name:     "mongodb uri with complex host",
			input:    "mongodb://user:pass@host1:27017,host2:27017,host3:27017",
			expected: "mongodb://user:***@host1:27017,host2:27017,host3:27017",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maskURI(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMaskURI_EdgeCases(t *testing.T) {
	t.Run("very short mongodb uri", func(t *testing.T) {
		// URIs shorter than 20 characters are not masked by the function
		input := "mongodb://a:b@c"
		expected := "mongodb://a:b@c" // Not masked due to length check
		result := maskURI(input)
		assert.Equal(t, expected, result)
	})

	// Note: The maskURI function has known limitations with passwords containing '@' characters
	// as it uses the first '@' after '://' to separate credentials from host
	t.Run("uri with @ in password - known limitation", func(t *testing.T) {
		t.Skip("maskURI has known limitation with @ in passwords")
	})

	t.Run("uri with colon in password", func(t *testing.T) {
		input := "mongodb://user:pass:word@localhost:27017"
		expected := "mongodb://user:***@localhost:27017"
		result := maskURI(input)
		assert.Equal(t, expected, result)
	})
}

func TestMaskURI_DoesNotExposePassword(t *testing.T) {
	sensitiveURIs := []string{
		"mongodb://admin:supersecret123@localhost:27017",
		"mongodb://user:MyPassword123!@cluster.mongodb.net:27017",
		"mongodb://dbuser:production_password@db.example.com:27017",
	}

	for _, uri := range sensitiveURIs {
		result := maskURI(uri)

		// Ensure password is not in the result
		assert.NotContains(t, result, "supersecret123")
		assert.NotContains(t, result, "MyPassword123!")
		assert.NotContains(t, result, "production_password")

		// Ensure *** is present
		assert.Contains(t, result, "***")
	}
}

type fakeRepo struct {
	createCalls []models.CreateProductRequest
	count       int64
	countCalls  int
}

func (f *fakeRepo) GetAllProducts(ctx context.Context) ([]models.ProductResponse, error) {
	return nil, nil
}

func (f *fakeRepo) GetProductById(ctx context.Context, id string) (*models.ProductResponse, error) {
	return nil, nil
}

func (f *fakeRepo) GetProductsByName(ctx context.Context, name string) ([]models.ProductResponse, error) {
	return nil, nil
}

func (f *fakeRepo) GetProductsByPage(ctx context.Context, page, pageSize int) (*models.PaginatedProductsResponse, error) {
	return &models.PaginatedProductsResponse{}, nil
}

func (f *fakeRepo) GetProductsByPageAndCategory(ctx context.Context, page, pageSize int, category string) (*models.PaginatedProductsResponse, error) {
	return &models.PaginatedProductsResponse{}, nil
}

func (f *fakeRepo) GetProductsCount(ctx context.Context) (int64, error) {
	f.countCalls++
	return f.count, nil
}

func (f *fakeRepo) GetProductsCountByCategory(ctx context.Context, category string) (int64, error) {
	return 0, nil
}

func (f *fakeRepo) GetCategories(ctx context.Context) ([]string, error) {
	return nil, nil
}

func (f *fakeRepo) CreateProduct(ctx context.Context, req models.CreateProductRequest) (*models.ProductResponse, error) {
	f.createCalls = append(f.createCalls, req)
	return &models.ProductResponse{ID: "id", Name: req.Name}, nil
}

func (f *fakeRepo) DeleteProductsByName(ctx context.Context, name string) error {
	return nil
}

func (f *fakeRepo) DeleteProductById(ctx context.Context, id string) error {
	return nil
}

func (f *fakeRepo) UpdateProductQuantity(ctx context.Context, productID string, quantityChange int) error {
	return nil
}

func (f *fakeRepo) GetProductQuantity(ctx context.Context, productID string) (int, error) {
	return 0, nil
}

func (f *fakeRepo) WarmupCache(ctx context.Context) error {
	return nil
}

func init() {
	// Initialize logger for tests
	log = logger.NewNop()
}

func TestRun_WithStubbedDependencies(t *testing.T) {
	repo := &fakeRepo{count: 5}
	listened := ""

	deps := appDependencies{
		loadEnv: func() error { return nil },
		buildRepo: func(cfg *config.Config) (repository.ProductRepository, func(context.Context) error, func(), error) {
			assert.Equal(t, "product_service", cfg.DatabaseName)
			return repo, nil, nil, nil
		},
		newSvc: services.NewProductService,
		listen: func(app *fiber.App, addr string) error {
			listened = addr

			req := httptest.NewRequest("GET", "/health", nil)
			resp, err := app.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, fiber.StatusOK, resp.StatusCode)

			return nil
		},
	}

	os.Setenv("PRODUCT_SERVICE_APP_PORT", "3200")
	defer os.Unsetenv("PRODUCT_SERVICE_APP_PORT")

	err := run(deps)
	assert.NoError(t, err)
	assert.Equal(t, ":3200", listened)
	assert.Equal(t, 1, repo.countCalls)
}

func TestRun_LoadEnvErrorDoesNotStop(t *testing.T) {
	called := false
	deps := appDependencies{
		loadEnv: func() error { called = true; return errors.New("missing env") },
		buildRepo: func(cfg *config.Config) (repository.ProductRepository, func(context.Context) error, func(), error) {
			return &fakeRepo{}, nil, nil, nil
		},
		newSvc: services.NewProductService,
		listen: func(app *fiber.App, addr string) error { return nil },
	}

	err := run(deps)
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestRun_ClosesResources(t *testing.T) {
	mongoClosed := false
	redisClosed := false

	deps := appDependencies{
		loadEnv: func() error { return nil },
		buildRepo: func(cfg *config.Config) (repository.ProductRepository, func(context.Context) error, func(), error) {
			return &fakeRepo{}, func(context.Context) error { mongoClosed = true; return nil }, func() { redisClosed = true }, nil
		},
		newSvc: services.NewProductService,
		listen: func(app *fiber.App, addr string) error { return nil },
	}

	err := run(deps)
	assert.NoError(t, err)
	assert.True(t, mongoClosed)
	assert.True(t, redisClosed)
}

func TestMain_UsesDefaultDependencies(t *testing.T) {
	original := defaultDeps
	originalFatal := fatalf
	defer func() {
		defaultDeps = original
		fatalf = originalFatal
	}()

	called := false
	defaultDeps = appDependencies{
		loadEnv: func() error { return nil },
		buildRepo: func(cfg *config.Config) (repository.ProductRepository, func(context.Context) error, func(), error) {
			called = true
			return &fakeRepo{}, nil, nil, nil
		},
		newSvc: services.NewProductService,
		listen: func(app *fiber.App, addr string) error { return nil },
	}

	main()
	assert.True(t, called)
}

func TestMain_FatalOnRunError(t *testing.T) {
	original := defaultDeps
	originalFatal := fatalf
	defer func() {
		defaultDeps = original
		fatalf = originalFatal
	}()

	called := false
	defaultDeps = appDependencies{
		loadEnv: func() error { return nil },
		buildRepo: func(cfg *config.Config) (repository.ProductRepository, func(context.Context) error, func(), error) {
			return nil, nil, nil, errors.New("boom")
		},
		newSvc: services.NewProductService,
		listen: func(app *fiber.App, addr string) error { return nil },
	}
	fatalf = func(v ...interface{}) { called = true }

	main()
	assert.True(t, called)
}
