package repository

import (
	"context"
	"testing"
	"time"

	"product-service/internal/model"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type fakeCollection struct {
	products []models.Product
}

func (f *fakeCollection) Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (*mongo.Cursor, error) {
	filtered := f.filterProducts(filter)
	var skip, limit int64
	if len(opts) > 0 && opts[0] != nil {
		if opts[0].Skip != nil {
			skip = *opts[0].Skip
		}
		if opts[0].Limit != nil {
			limit = *opts[0].Limit
		}
	}
	start := int(skip)
	end := len(filtered)
	if limit > 0 && start+int(limit) < end {
		end = start + int(limit)
	}
	if start > len(filtered) {
		filtered = []models.Product{}
	} else {
		filtered = filtered[start:end]
	}
	docs := make([]interface{}, len(filtered))
	for i, p := range filtered {
		docs[i] = p
	}
	return mongo.NewCursorFromDocuments(docs, nil, nil)
}

func (f *fakeCollection) CountDocuments(ctx context.Context, filter interface{}, opts ...*options.CountOptions) (int64, error) {
	return int64(len(f.filterProducts(filter))), nil
}

func (f *fakeCollection) Distinct(ctx context.Context, field string, filter interface{}, opts ...*options.DistinctOptions) ([]interface{}, error) {
	set := make(map[string]struct{})
	for _, p := range f.filterProducts(filter) {
		set[p.Category] = struct{}{}
	}
	out := make([]interface{}, 0, len(set))
	for key := range set {
		out = append(out, key)
	}
	return out, nil
}

func (f *fakeCollection) InsertOne(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	return &mongo.InsertOneResult{InsertedID: primitive.NewObjectID()}, nil
}

func (f *fakeCollection) DeleteMany(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	return &mongo.DeleteResult{DeletedCount: 1}, nil
}

func (f *fakeCollection) DeleteOne(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	return &mongo.DeleteResult{DeletedCount: 1}, nil
}

func (f *fakeCollection) UpdateOne(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	return &mongo.UpdateResult{MatchedCount: 1, ModifiedCount: 1}, nil
}

func (f *fakeCollection) FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) *mongo.SingleResult {
	filtered := f.filterProducts(filter)
	if len(filtered) == 0 {
		return mongo.NewSingleResultFromDocument(bson.D{}, mongo.ErrNoDocuments, nil)
	}
	return mongo.NewSingleResultFromDocument(filtered[0], nil, nil)
}

func (f *fakeCollection) filterProducts(filter interface{}) []models.Product {
	if filter == nil {
		return f.products
	}
	switch v := filter.(type) {
	case bson.M:
		if category, ok := v["category"].(string); ok {
			var res []models.Product
			for _, p := range f.products {
				if p.Category == category {
					res = append(res, p)
				}
			}
			return res
		}
		if name, ok := v["name"].(string); ok {
			var res []models.Product
			for _, p := range f.products {
				if p.Name == name {
					res = append(res, p)
				}
			}
			return res
		}
		if id, ok := v["_id"].(primitive.ObjectID); ok {
			for _, p := range f.products {
				if p.ID == id {
					return []models.Product{p}
				}
			}
			return nil
		}
	}
	return f.products
}

func TestGetProductsByPage_WithFakeCollection(t *testing.T) {
	ctx := context.Background()
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	now := time.Now()
	products := []models.Product{
		{ID: primitive.NewObjectID(), Name: "A", Price: 10, Quantity: 2, Category: "Cats", DateCreated: now, DateUpdated: now},
		{ID: primitive.NewObjectID(), Name: "B", Price: 11, Quantity: 3, Category: "Dogs", DateCreated: now, DateUpdated: now},
		{ID: primitive.NewObjectID(), Name: "C", Price: 12, Quantity: 4, Category: "Cats", DateCreated: now, DateUpdated: now},
	}

	repo := &productRepository{collection: &fakeCollection{products: products}, redis: rdb}
	resp, err := repo.GetProductsByPage(ctx, 1, 2)
	require.NoError(t, err)
	require.Equal(t, int64(len(products)), resp.TotalCount)
	require.Len(t, resp.Products, 2)
	require.True(t, mr.Exists("productsPage:1:2"))
}

func TestGetProductsByPageAndCategory_WithFakeCollection(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	products := []models.Product{
		{ID: primitive.NewObjectID(), Name: "A", Price: 10, Quantity: 2, Category: "Cats", DateCreated: now, DateUpdated: now},
		{ID: primitive.NewObjectID(), Name: "B", Price: 11, Quantity: 3, Category: "Dogs", DateCreated: now, DateUpdated: now},
		{ID: primitive.NewObjectID(), Name: "C", Price: 12, Quantity: 4, Category: "Cats", DateCreated: now, DateUpdated: now},
	}

	repo := &productRepository{collection: &fakeCollection{products: products}}
	resp, err := repo.GetProductsByPageAndCategory(ctx, 1, 2, "Cats")
	require.NoError(t, err)
	require.Equal(t, int64(2), resp.TotalCount)
	require.Len(t, resp.Products, 2)
}
