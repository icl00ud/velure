package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"product-service/internal/metrics"
	"product-service/internal/models"
	"product-service/internal/repository"
)

type ProductService interface {
	GetAllProducts(ctx context.Context) ([]models.ProductResponse, error)
	GetProductsByName(ctx context.Context, name string) ([]models.ProductResponse, error)
	GetProductsByPage(ctx context.Context, page, pageSize int) (*models.PaginatedProductsResponse, error)
	GetProductsByPageAndCategory(ctx context.Context, page, pageSize int, category string) (*models.PaginatedProductsResponse, error)
	GetProductsCount(ctx context.Context) (*models.CountResponse, error)
	GetCategories(ctx context.Context) ([]string, error)
	CreateProduct(ctx context.Context, product models.CreateProductRequest) (*models.ProductResponse, error)
	DeleteProductsByName(ctx context.Context, name string) error
	DeleteProductById(ctx context.Context, id string) error
	UpdateProductQuantity(ctx context.Context, productID string, quantityChange int) error
	SyncProductCatalogMetric(ctx context.Context)
}

type productService struct {
	repo repository.ProductRepository
}

func NewProductService(repo repository.ProductRepository) ProductService {
	return &productService{
		repo: repo,
	}
}

func (s *productService) GetAllProducts(ctx context.Context) ([]models.ProductResponse, error) {
	start := time.Now()
	defer func() {
		metrics.ProductOperationDuration.WithLabelValues("get_all").Observe(time.Since(start).Seconds())
	}()

	metrics.ProductQueries.WithLabelValues("get_all").Inc()
	results, err := s.repo.GetAllProducts(ctx)
	if err != nil {
		metrics.Errors.WithLabelValues("database").Inc()
		return nil, err
	}

	metrics.SearchResultsReturned.Observe(float64(len(results)))
	return results, nil
}

func (s *productService) GetProductsByName(ctx context.Context, name string) ([]models.ProductResponse, error) {
	start := time.Now()
	defer func() {
		metrics.ProductOperationDuration.WithLabelValues("get_by_name").Observe(time.Since(start).Seconds())
	}()

	metrics.ProductQueries.WithLabelValues("get_by_name").Inc()
	metrics.ProductSearches.WithLabelValues("by_name").Inc()

	results, err := s.repo.GetProductsByName(ctx, name)
	if err != nil {
		metrics.Errors.WithLabelValues("database").Inc()
		return nil, err
	}

	metrics.SearchResultsReturned.Observe(float64(len(results)))
	return results, nil
}

func (s *productService) GetProductsByPage(ctx context.Context, page, pageSize int) (*models.PaginatedProductsResponse, error) {
	start := time.Now()
	defer func() {
		metrics.ProductOperationDuration.WithLabelValues("get_by_page").Observe(time.Since(start).Seconds())
	}()

	metrics.ProductQueries.WithLabelValues("get_by_page").Inc()
	metrics.ProductSearches.WithLabelValues("paginated").Inc()

	result, err := s.repo.GetProductsByPage(ctx, page, pageSize)
	if err != nil {
		metrics.Errors.WithLabelValues("database").Inc()
		return nil, err
	}

	return result, nil
}

func (s *productService) GetProductsByPageAndCategory(ctx context.Context, page, pageSize int, category string) (*models.PaginatedProductsResponse, error) {
	start := time.Now()
	defer func() {
		metrics.ProductOperationDuration.WithLabelValues("get_by_category").Observe(time.Since(start).Seconds())
	}()

	metrics.ProductQueries.WithLabelValues("get_by_category").Inc()
	metrics.ProductSearches.WithLabelValues("by_category").Inc()

	result, err := s.repo.GetProductsByPageAndCategory(ctx, page, pageSize, category)
	if err != nil {
		metrics.Errors.WithLabelValues("database").Inc()
		return nil, err
	}

	return result, nil
}

func (s *productService) GetProductsCount(ctx context.Context) (*models.CountResponse, error) {
	metrics.ProductQueries.WithLabelValues("get_count").Inc()

	count, err := s.repo.GetProductsCount(ctx)
	if err != nil {
		metrics.Errors.WithLabelValues("database").Inc()
		return nil, err
	}

	metrics.CurrentProductCount.Set(float64(count))
	return &models.CountResponse{Count: count}, nil
}

func (s *productService) GetCategories(ctx context.Context) ([]string, error) {
	metrics.CategoryQueries.Inc()

	results, err := s.repo.GetCategories(ctx)
	if err != nil {
		metrics.Errors.WithLabelValues("database").Inc()
		return nil, err
	}

	return results, nil
}

func (s *productService) CreateProduct(ctx context.Context, product models.CreateProductRequest) (*models.ProductResponse, error) {
	start := time.Now()
	defer func() {
		metrics.ProductOperationDuration.WithLabelValues("create").Observe(time.Since(start).Seconds())
	}()

	result, err := s.repo.CreateProduct(ctx, product)
	if err != nil {
		metrics.ProductMutations.WithLabelValues("create", "failure").Inc()
		metrics.Errors.WithLabelValues("database").Inc()
		return nil, err
	}

	metrics.ProductMutations.WithLabelValues("create", "success").Inc()
	s.SyncProductCatalogMetric(ctx)
	return result, nil
}

func (s *productService) DeleteProductsByName(ctx context.Context, name string) error {
	start := time.Now()
	defer func() {
		metrics.ProductOperationDuration.WithLabelValues("delete").Observe(time.Since(start).Seconds())
	}()

	err := s.repo.DeleteProductsByName(ctx, name)
	if err != nil {
		metrics.ProductMutations.WithLabelValues("delete", "failure").Inc()
		metrics.Errors.WithLabelValues("database").Inc()
		return err
	}

	metrics.ProductMutations.WithLabelValues("delete", "success").Inc()
	s.SyncProductCatalogMetric(ctx)
	return nil
}

func (s *productService) DeleteProductById(ctx context.Context, id string) error {
	start := time.Now()
	defer func() {
		metrics.ProductOperationDuration.WithLabelValues("delete").Observe(time.Since(start).Seconds())
	}()

	err := s.repo.DeleteProductById(ctx, id)
	if err != nil {
		metrics.ProductMutations.WithLabelValues("delete", "failure").Inc()
		metrics.Errors.WithLabelValues("database").Inc()
		return err
	}

	metrics.ProductMutations.WithLabelValues("delete", "success").Inc()
	s.SyncProductCatalogMetric(ctx)
	return nil
}

func (s *productService) UpdateProductQuantity(ctx context.Context, productID string, quantityChange int) error {
	start := time.Now()
	var status string
	defer func() {
		metrics.ProductOperationDuration.WithLabelValues("update_quantity").Observe(time.Since(start).Seconds())
		metrics.InventoryUpdates.WithLabelValues(status).Inc()
	}()

	// Validate that the update won't result in negative quantity
	currentQuantity, err := s.repo.GetProductQuantity(ctx, productID)
	if err != nil {
		status = "failure"
		metrics.Errors.WithLabelValues("database").Inc()
		return err
	}

	newQuantity := currentQuantity + quantityChange
	if newQuantity < 0 {
		status = "insufficient_stock"
		metrics.Errors.WithLabelValues("validation").Inc()
		return fmt.Errorf("insufficient stock: current quantity is %d, cannot deduct %d", currentQuantity, -quantityChange)
	}

	err = s.repo.UpdateProductQuantity(ctx, productID, quantityChange)
	if err != nil {
		status = "failure"
		metrics.Errors.WithLabelValues("database").Inc()
		return err
	}

	status = "success"
	return nil
}

func (s *productService) SyncProductCatalogMetric(ctx context.Context) {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	count, err := s.repo.GetProductsCount(ctx)
	if err != nil {
		log.Printf("failed to sync product catalog metric: %v", err)
		return
	}
	metrics.CurrentProductCount.Set(float64(count))
}
