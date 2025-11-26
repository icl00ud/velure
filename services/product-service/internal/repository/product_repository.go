package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"product-service/internal/metrics"
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
	GetProductsByPage(ctx context.Context, page, pageSize int) (*models.PaginatedProductsResponse, error)
	GetProductsByPageAndCategory(ctx context.Context, page, pageSize int, category string) (*models.PaginatedProductsResponse, error)
	GetProductsCount(ctx context.Context) (int64, error)
	GetProductsCountByCategory(ctx context.Context, category string) (int64, error)
	GetCategories(ctx context.Context) ([]string, error)
	CreateProduct(ctx context.Context, product models.CreateProductRequest) (*models.ProductResponse, error)
	DeleteProductsByName(ctx context.Context, name string) error
	DeleteProductById(ctx context.Context, id string) error
	UpdateProductQuantity(ctx context.Context, productID string, quantityChange int) error
	GetProductQuantity(ctx context.Context, productID string) (int, error)
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
		metrics.CacheHits.Inc()
		var products []models.ProductResponse
		if err := json.Unmarshal([]byte(cached), &products); err == nil {
			return products, nil
		}
	}
	metrics.CacheMisses.Inc()

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

func (r *productRepository) GetProductsByPage(ctx context.Context, page, pageSize int) (*models.PaginatedProductsResponse, error) {
	cacheKey := fmt.Sprintf("productsPage:%d:%d", page, pageSize)

	// Try to get from cache
	cached, err := r.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		metrics.CacheHits.Inc()
		var response models.PaginatedProductsResponse
		if err := json.Unmarshal([]byte(cached), &response); err == nil {
			return &response, nil
		}
	}
	metrics.CacheMisses.Inc()

	// Get total count
	totalCount, err := r.GetProductsCount(ctx)
	if err != nil {
		return nil, err
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

	productResponses := make([]models.ProductResponse, len(products))
	for i, product := range products {
		productResponses[i] = r.toProductResponse(product)
	}

	totalPages := int(totalCount) / pageSize
	if int(totalCount)%pageSize != 0 {
		totalPages++
	}

	response := &models.PaginatedProductsResponse{
		Products:   productResponses,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}

	// Cache the result
	if data, err := json.Marshal(response); err == nil {
		r.redis.Set(ctx, cacheKey, data, time.Hour)
	}

	return response, nil
}

func (r *productRepository) GetProductsByPageAndCategory(ctx context.Context, page, pageSize int, category string) (*models.PaginatedProductsResponse, error) {
	// Get total count for this category
	totalCount, err := r.GetProductsCountByCategory(ctx, category)
	if err != nil {
		return nil, err
	}

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

	productResponses := make([]models.ProductResponse, len(products))
	for i, product := range products {
		productResponses[i] = r.toProductResponse(product)
	}

	totalPages := int(totalCount) / pageSize
	if int(totalCount)%pageSize != 0 {
		totalPages++
	}

	return &models.PaginatedProductsResponse{
		Products:   productResponses,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func (r *productRepository) GetProductsCount(ctx context.Context) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{})
}

func (r *productRepository) GetProductsCountByCategory(ctx context.Context, category string) (int64, error) {
	filter := bson.M{"category": category}
	return r.collection.CountDocuments(ctx, filter)
}

func (r *productRepository) GetCategories(ctx context.Context) ([]string, error) {
	cacheKey := "productCategories"

	// Try to get from cache
	cached, err := r.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		metrics.CacheHits.Inc()
		var categories []string
		if err := json.Unmarshal([]byte(cached), &categories); err == nil {
			return categories, nil
		}
	}
	metrics.CacheMisses.Inc()

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
		ID:          primitive.NewObjectID(),
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Category:    req.Category,
		Quantity:    req.Quantity,
		Images:      req.Images,
		Dimensions:  req.Dimensions,
		Brand:       req.Brand,
		Colors:      req.Colors,
		SKU:         req.SKU,
		DateCreated: time.Now(),
		DateUpdated: time.Now(),
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
		ID:          product.ID.Hex(),
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Rating:      product.Rating,
		Category:    product.Category,
		Quantity:    product.Quantity,
		Images:      product.Images,
		Dimensions:  product.Dimensions,
		Brand:       product.Brand,
		Colors:      product.Colors,
		SKU:         product.SKU,
		DateCreated: product.DateCreated,
		DateUpdated: product.DateUpdated,
	}
}

func (r *productRepository) UpdateProductQuantity(ctx context.Context, productID string, quantityChange int) error {
	objectID, err := primitive.ObjectIDFromHex(productID)
	if err != nil {
		return fmt.Errorf("invalid product ID: %w", err)
	}

	// Use $inc to atomically increment/decrement the quantity
	update := bson.M{
		"$inc": bson.M{"quantity": quantityChange},
		"$set": bson.M{"dt_updated": time.Now()},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		return fmt.Errorf("failed to update quantity: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("product not found")
	}

	// Clear all product-related caches
	r.clearProductCaches(ctx)

	return nil
}

func (r *productRepository) GetProductQuantity(ctx context.Context, productID string) (int, error) {
	objectID, err := primitive.ObjectIDFromHex(productID)
	if err != nil {
		return 0, fmt.Errorf("invalid product ID: %w", err)
	}

	var product models.Product
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&product)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return 0, fmt.Errorf("product not found")
		}
		return 0, fmt.Errorf("failed to get product: %w", err)
	}

	return product.Quantity, nil
}

func (r *productRepository) clearProductCaches(ctx context.Context) {
	// Clear all products cache
	r.redis.Del(ctx, "allProducts")

	// Clear categories cache
	r.redis.Del(ctx, "productCategories")

	// Clear all pagination caches (pattern delete)
	// Note: This is a simple approach. For production, consider using Redis SCAN
	keys, err := r.redis.Keys(ctx, "productsPage:*").Result()
	if err == nil && len(keys) > 0 {
		r.redis.Del(ctx, keys...)
	}
}
