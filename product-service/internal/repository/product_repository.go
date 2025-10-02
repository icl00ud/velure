package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"product-service/internal/models"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ProductRepository interface {
	GetAllProducts(ctx context.Context) ([]models.ProductResponse, error)
	GetProductsByName(ctx context.Context, name string) ([]models.ProductResponse, error)
	GetProductsByPage(ctx context.Context, page, pageSize int) ([]models.ProductResponse, error)
	GetProductsByPageAndCategory(ctx context.Context, page, pageSize int, category string) ([]models.ProductResponse, error)
	GetProductsCount(ctx context.Context) (int64, error)
	GetCategories(ctx context.Context) ([]string, error)
	CreateProduct(ctx context.Context, product models.CreateProductRequest) (*models.ProductResponse, error)
	DeleteProductsByName(ctx context.Context, name string) error
	DeleteProductById(ctx context.Context, id string) error
}

type productRepository struct {
	collection *mongo.Collection
	redis      *redis.Client
}

func NewProductRepository(db *mongo.Database, redisClient *redis.Client) ProductRepository {
	return &productRepository{
		collection: db.Collection("products"),
		redis:      redisClient,
	}
}

func (r *productRepository) GetAllProducts(ctx context.Context) ([]models.ProductResponse, error) {
	cacheKey := "allProducts"

	// Try to get from cache
	cached, err := r.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var products []models.ProductResponse
		if err := json.Unmarshal([]byte(cached), &products); err == nil {
			return products, nil
		}
	}

	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var products []models.Product
	if err := cursor.All(ctx, &products); err != nil {
		return nil, err
	}

	response := make([]models.ProductResponse, len(products))
	for i, product := range products {
		response[i] = r.toProductResponse(product)
	}

	// Cache the result
	if data, err := json.Marshal(response); err == nil {
		r.redis.Set(ctx, cacheKey, data, time.Hour)
	}

	return response, nil
}

func (r *productRepository) GetProductsByName(ctx context.Context, name string) ([]models.ProductResponse, error) {
	filter := bson.M{"name": name}
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var products []models.Product
	if err := cursor.All(ctx, &products); err != nil {
		return nil, err
	}

	response := make([]models.ProductResponse, len(products))
	for i, product := range products {
		response[i] = r.toProductResponse(product)
	}

	return response, nil
}

func (r *productRepository) GetProductsByPage(ctx context.Context, page, pageSize int) ([]models.ProductResponse, error) {
	cacheKey := fmt.Sprintf("productsPage:%d:%d", page, pageSize)

	// Try to get from cache
	cached, err := r.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var products []models.ProductResponse
		if err := json.Unmarshal([]byte(cached), &products); err == nil {
			return products, nil
		}
	}

	skip := int64((page - 1) * pageSize)
	limit := int64(pageSize)

	opts := options.Find().SetSkip(skip).SetLimit(limit)
	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var products []models.Product
	if err := cursor.All(ctx, &products); err != nil {
		return nil, err
	}

	response := make([]models.ProductResponse, len(products))
	for i, product := range products {
		response[i] = r.toProductResponse(product)
	}

	// Cache the result
	if data, err := json.Marshal(response); err == nil {
		r.redis.Set(ctx, cacheKey, data, time.Hour)
	}

	return response, nil
}

func (r *productRepository) GetProductsByPageAndCategory(ctx context.Context, page, pageSize int, category string) ([]models.ProductResponse, error) {
	skip := int64((page - 1) * pageSize)
	limit := int64(pageSize)

	filter := bson.M{"category": category}
	opts := options.Find().SetSkip(skip).SetLimit(limit)
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var products []models.Product
	if err := cursor.All(ctx, &products); err != nil {
		return nil, err
	}

	response := make([]models.ProductResponse, len(products))
	for i, product := range products {
		response[i] = r.toProductResponse(product)
	}

	return response, nil
}

func (r *productRepository) GetProductsCount(ctx context.Context) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{})
}

func (r *productRepository) GetCategories(ctx context.Context) ([]string, error) {
	cacheKey := "productCategories"

	// Try to get from cache
	cached, err := r.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var categories []string
		if err := json.Unmarshal([]byte(cached), &categories); err == nil {
			return categories, nil
		}
	}

	// Get distinct categories from MongoDB
	categories, err := r.collection.Distinct(ctx, "category", bson.M{})
	if err != nil {
		return nil, err
	}

	// Convert interface{} to []string
	result := make([]string, 0, len(categories))
	for _, cat := range categories {
		if catStr, ok := cat.(string); ok && catStr != "" {
			result = append(result, catStr)
		}
	}

	// Cache the result for 1 hour
	if data, err := json.Marshal(result); err == nil {
		r.redis.Set(ctx, cacheKey, data, time.Hour)
	}

	return result, nil
}

func (r *productRepository) CreateProduct(ctx context.Context, req models.CreateProductRequest) (*models.ProductResponse, error) {
	product := models.Product{
		ID:                primitive.NewObjectID(),
		Name:              req.Name,
		Description:       req.Description,
		Price:             req.Price,
		Category:          req.Category,
		Disponibility:     req.Disponibility,
		QuantityWarehouse: req.QuantityWarehouse,
		Images:            req.Images,
		Dimensions:        req.Dimensions,
		Brand:             req.Brand,
		Colors:            req.Colors,
		SKU:               req.SKU,
		DateCreated:       time.Now(),
		DateUpdated:       time.Now(),
	}

	_, err := r.collection.InsertOne(ctx, product)
	if err != nil {
		return nil, err
	}

	// Clear cache
	r.redis.Del(ctx, "allProducts")

	response := r.toProductResponse(product)
	return &response, nil
}

func (r *productRepository) DeleteProductsByName(ctx context.Context, name string) error {
	_, err := r.collection.DeleteMany(ctx, bson.M{"name": name})
	if err != nil {
		return err
	}

	// Clear cache
	r.redis.Del(ctx, "allProducts")
	return nil
}

func (r *productRepository) DeleteProductById(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return err
	}

	// Clear cache
	r.redis.Del(ctx, "allProducts")
	return nil
}

func (r *productRepository) toProductResponse(product models.Product) models.ProductResponse {
	return models.ProductResponse{
		ID:                product.ID.Hex(),
		Name:              product.Name,
		Description:       product.Description,
		Price:             product.Price,
		Category:          product.Category,
		Disponibility:     product.Disponibility,
		QuantityWarehouse: product.QuantityWarehouse,
		Images:            product.Images,
		Dimensions:        product.Dimensions,
		Brand:             product.Brand,
		Colors:            product.Colors,
		SKU:               product.SKU,
		DateCreated:       product.DateCreated,
		DateUpdated:       product.DateUpdated,
	}
}
