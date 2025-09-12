package services

import (
	"context"

	"product-service/internal/models"
	"product-service/internal/repository"
)

type ProductService interface {
	GetAllProducts(ctx context.Context) ([]models.ProductResponse, error)
	GetProductsByName(ctx context.Context, name string) ([]models.ProductResponse, error)
	GetProductsByPage(ctx context.Context, page, pageSize int) ([]models.ProductResponse, error)
	GetProductsByPageAndCategory(ctx context.Context, page, pageSize int, category string) ([]models.ProductResponse, error)
	GetProductsCount(ctx context.Context) (*models.CountResponse, error)
	CreateProduct(ctx context.Context, product models.CreateProductRequest) (*models.ProductResponse, error)
	DeleteProductsByName(ctx context.Context, name string) error
	DeleteProductById(ctx context.Context, id string) error
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
	return s.repo.GetAllProducts(ctx)
}

func (s *productService) GetProductsByName(ctx context.Context, name string) ([]models.ProductResponse, error) {
	return s.repo.GetProductsByName(ctx, name)
}

func (s *productService) GetProductsByPage(ctx context.Context, page, pageSize int) ([]models.ProductResponse, error) {
	return s.repo.GetProductsByPage(ctx, page, pageSize)
}

func (s *productService) GetProductsByPageAndCategory(ctx context.Context, page, pageSize int, category string) ([]models.ProductResponse, error) {
	return s.repo.GetProductsByPageAndCategory(ctx, page, pageSize, category)
}

func (s *productService) GetProductsCount(ctx context.Context) (*models.CountResponse, error) {
	count, err := s.repo.GetProductsCount(ctx)
	if err != nil {
		return nil, err
	}
	return &models.CountResponse{Count: count}, nil
}

func (s *productService) CreateProduct(ctx context.Context, product models.CreateProductRequest) (*models.ProductResponse, error) {
	return s.repo.CreateProduct(ctx, product)
}

func (s *productService) DeleteProductsByName(ctx context.Context, name string) error {
	return s.repo.DeleteProductsByName(ctx, name)
}

func (s *productService) DeleteProductById(ctx context.Context, id string) error {
	return s.repo.DeleteProductById(ctx, id)
}
