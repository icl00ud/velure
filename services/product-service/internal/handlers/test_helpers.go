package handlers

import (
	"context"

	"github.com/gofiber/fiber/v2"

	"product-service/internal/models"
	"product-service/internal/services"
)

// stubProductService implements services.ProductService for handler error-path tests.
type stubProductService struct {
	err error
}

func (s *stubProductService) SyncProductCatalogMetric(ctx context.Context) {}
func (s *stubProductService) GetAllProducts(ctx context.Context) ([]models.ProductResponse, error) {
	return nil, s.err
}
func (s *stubProductService) GetProductById(ctx context.Context, id string) (*models.ProductResponse, error) {
	return nil, s.err
}
func (s *stubProductService) GetProductsByName(ctx context.Context, name string) ([]models.ProductResponse, error) {
	return nil, s.err
}
func (s *stubProductService) GetProductsByPage(ctx context.Context, page, pageSize int) (*models.PaginatedProductsResponse, error) {
	return nil, s.err
}
func (s *stubProductService) GetProductsByPageAndCategory(ctx context.Context, page, pageSize int, category string) (*models.PaginatedProductsResponse, error) {
	return nil, s.err
}
func (s *stubProductService) GetProductsByPageAndCategoryFromCache(ctx context.Context, page, pageSize int, category string) (*models.PaginatedProductsResponse, error) {
	return nil, s.err
}
func (s *stubProductService) GetProductsCount(ctx context.Context) (*models.CountResponse, error) {
	return nil, s.err
}
func (s *stubProductService) GetProductsCountByCategory(ctx context.Context, category string) (int64, error) {
	return 0, s.err
}
func (s *stubProductService) GetCategories(ctx context.Context) ([]string, error) { return nil, s.err }
func (s *stubProductService) CreateProduct(ctx context.Context, req models.CreateProductRequest) (*models.ProductResponse, error) {
	return nil, s.err
}
func (s *stubProductService) DeleteProductsByName(ctx context.Context, name string) error {
	return s.err
}
func (s *stubProductService) DeleteProductById(ctx context.Context, id string) error { return s.err }
func (s *stubProductService) UpdateProductQuantity(ctx context.Context, productID string, quantityChange int) error {
	return s.err
}
func (s *stubProductService) GetProductQuantity(ctx context.Context, productID string) (int, error) {
	return 0, s.err
}

// Helper to initialize Fiber app with handler for tests.
func setupTestApp() (*fiber.App, *ProductHandler) {
	app := fiber.New()
	handler := &ProductHandler{}
	return app, handler
}

// Ensure stub satisfies interface
var _ services.ProductService = (*stubProductService)(nil)
