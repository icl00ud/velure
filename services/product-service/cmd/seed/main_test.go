package main

import (
	"context"
	"errors"
	"math/rand"
	"strings"
	"testing"

	"product-service/internal/config"
	"product-service/internal/models"
	"product-service/internal/repository"

	"github.com/stretchr/testify/assert"
)

type recordingRepo struct {
	createCalls []models.CreateProductRequest
	createErrs  []error
}

func (r *recordingRepo) GetAllProducts(ctx context.Context) ([]models.ProductResponse, error) {
	return nil, nil
}
func (r *recordingRepo) GetProductById(ctx context.Context, id string) (*models.ProductResponse, error) {
	return nil, nil
}
func (r *recordingRepo) GetProductsByName(ctx context.Context, name string) ([]models.ProductResponse, error) {
	return nil, nil
}
func (r *recordingRepo) GetProductsByPage(ctx context.Context, page, pageSize int) (*models.PaginatedProductsResponse, error) {
	return &models.PaginatedProductsResponse{}, nil
}
func (r *recordingRepo) GetProductsByPageAndCategory(ctx context.Context, page, pageSize int, category string) (*models.PaginatedProductsResponse, error) {
	return &models.PaginatedProductsResponse{}, nil
}
func (r *recordingRepo) GetProductsCount(ctx context.Context) (int64, error) { return 0, nil }
func (r *recordingRepo) GetProductsCountByCategory(ctx context.Context, category string) (int64, error) {
	return 0, nil
}
func (r *recordingRepo) GetCategories(ctx context.Context) ([]string, error) { return nil, nil }
func (r *recordingRepo) CreateProduct(ctx context.Context, product models.CreateProductRequest) (*models.ProductResponse, error) {
	r.createCalls = append(r.createCalls, product)
	if len(r.createErrs) > 0 {
		err := r.createErrs[0]
		r.createErrs = r.createErrs[1:]
		if err != nil {
			return nil, err
		}
	}
	return &models.ProductResponse{ID: "id", Name: product.Name}, nil
}
func (r *recordingRepo) DeleteProductsByName(ctx context.Context, name string) error { return nil }
func (r *recordingRepo) DeleteProductById(ctx context.Context, id string) error      { return nil }
func (r *recordingRepo) UpdateProductQuantity(ctx context.Context, productID string, quantityChange int) error {
	return nil
}
func (r *recordingRepo) GetProductQuantity(ctx context.Context, productID string) (int, error) {
	return 0, nil
}
func (r *recordingRepo) WarmupCache(ctx context.Context) error { return nil }

func TestRunSeed_InsertsProducts(t *testing.T) {
	repo := &recordingRepo{
		createErrs: []error{errors.New("fail once"), nil, nil, nil, nil},
	}

	deps := seedDependencies{
		loadEnv: func() error { return nil },
		buildRepo: func(cfg *config.Config) (repository.ProductRepository, func(context.Context) error, func(), error) {
			return repo, nil, nil, nil
		},
		generate: func() []models.CreateProductRequest {
			return []models.CreateProductRequest{
				{Name: "A"}, {Name: "B"}, {Name: "C"}, {Name: "D"}, {Name: "E"},
			}
		},
	}

	err := runSeed(deps)
	assert.NoError(t, err)
	assert.Len(t, repo.createCalls, 5)
}

func TestRunSeed_RepoError(t *testing.T) {
	deps := seedDependencies{
		loadEnv: func() error { return nil },
		buildRepo: func(cfg *config.Config) (repository.ProductRepository, func(context.Context) error, func(), error) {
			return nil, nil, nil, errors.New("connect failed")
		},
		generate: generatePetProducts,
	}

	err := runSeed(deps)
	assert.Error(t, err)
}

func TestRunSeed_WithLoadEnvError(t *testing.T) {
	repo := &recordingRepo{}
	deps := seedDependencies{
		loadEnv: func() error { return errors.New("missing env") },
		buildRepo: func(cfg *config.Config) (repository.ProductRepository, func(context.Context) error, func(), error) {
			return repo, func(context.Context) error { return nil }, func() {}, nil
		},
		generate: func() []models.CreateProductRequest {
			return []models.CreateProductRequest{{Name: "Single", Quantity: 1, Price: 1.0, SKU: "S-1"}}
		},
	}

	err := runSeed(deps)
	assert.NoError(t, err)
	assert.Len(t, repo.createCalls, 1)
}

func TestSeedMainUsesDefaultDeps(t *testing.T) {
	original := defaultSeedDeps
	originalFatal := seedFatalf
	defer func() {
		defaultSeedDeps = original
		seedFatalf = originalFatal
	}()

	called := false
	defaultSeedDeps = seedDependencies{
		loadEnv: func() error { called = true; return nil },
		buildRepo: func(cfg *config.Config) (repository.ProductRepository, func(context.Context) error, func(), error) {
			return &recordingRepo{}, nil, nil, nil
		},
		generate: func() []models.CreateProductRequest { return nil },
	}

	main()
	assert.True(t, called)
}

func TestSeedMain_FatalOnError(t *testing.T) {
	original := defaultSeedDeps
	originalFatal := seedFatalf
	defer func() {
		defaultSeedDeps = original
		seedFatalf = originalFatal
	}()

	called := false
	defaultSeedDeps = seedDependencies{
		loadEnv: func() error { return nil },
		buildRepo: func(cfg *config.Config) (repository.ProductRepository, func(context.Context) error, func(), error) {
			return nil, nil, nil, errors.New("seed failure")
		},
		generate: generatePetProducts,
	}
	seedFatalf = func(v ...interface{}) { called = true }

	main()
	assert.True(t, called)
}

func TestGeneratePetProductsProducesInventory(t *testing.T) {
	products := generatePetProducts()
	assert.Equal(t, 16, len(products))

	for _, p := range products {
		assert.NotEmpty(t, p.Name)
		assert.True(t, p.Rating >= 3.5 && p.Rating <= 5.0)
		assert.True(t, p.Quantity >= 10)
		assert.Len(t, p.Images, 3)
		assert.NotEmpty(t, p.SKU)
	}
}

func TestRandomRatingRange(t *testing.T) {
	rand.Seed(1)
	for i := 0; i < 5; i++ {
		r := randomRating()
		assert.GreaterOrEqual(t, r, 3.5)
		assert.LessOrEqual(t, r, 5.0)
	}
}

func TestRandomQuantityRange(t *testing.T) {
	rand.Seed(2)
	for i := 0; i < 5; i++ {
		q := randomQuantity()
		assert.GreaterOrEqual(t, q, 10)
		assert.LessOrEqual(t, q, 110)
	}
}

func TestGenerateImagesFallback(t *testing.T) {
	rand.Seed(3)
	images := generateImages("Unknown", "Item")
	assert.Len(t, images, 3)
	for _, img := range images {
		assert.True(t, strings.HasPrefix(img, "https://"))
	}
}

func TestGenerateDimensionsDefaults(t *testing.T) {
	dims := generateDimensions("Brinquedos")
	assert.Equal(t, 10.0, dims.Height)
	assert.Equal(t, 10.0, dims.Width)
	assert.Equal(t, 15.0, dims.Length)

	fallback := generateDimensions("Nope")
	assert.Equal(t, 10.0, fallback.Height)
	assert.Equal(t, 10.0, fallback.Width)
	assert.Equal(t, 10.0, fallback.Length)
}

func TestGenerateColorsSelection(t *testing.T) {
	rand.Seed(4)
	colors := generateColors("Brinquedos")
	assert.True(t, len(colors) >= 1 && len(colors) <= 5)

	defaultColors := generateColors("Desconhecida")
	assert.Equal(t, []string{"Variadas"}, defaultColors)
}

func TestRandomStringUsesCharset(t *testing.T) {
	rand.Seed(5)
	val := randomString(12)
	assert.Len(t, val, 12)

	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	for _, ch := range val {
		assert.True(t, strings.ContainsRune(charset, ch))
	}
}
