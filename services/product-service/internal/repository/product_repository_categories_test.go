package repository

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestGetCategoriesCachesResults(t *testing.T) {
	ctx := context.Background()
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("cache and fetch", func(mt *mtest.T) {
		mr, err := miniredis.Run()
		require.NoError(mt, err)
		defer mr.Close()

		rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
		repo := &productRepository{collection: mt.Coll, redis: rdb}

		mt.AddMockResponses(mtest.CreateSuccessResponse(
			bson.E{Key: "values", Value: bson.A{"Dogs", "Cats"}},
		))

		cats, err := repo.GetCategories(ctx)
		require.NoError(mt, err)
		require.Equal(mt, []string{"Dogs", "Cats"}, cats)

		// Ensure cache was populated
		require.True(mt, mr.Exists("productCategories"))

		// Second call should hit cache and return same result
		cached, err := repo.GetCategories(ctx)
		require.NoError(mt, err)
		require.Equal(mt, cats, cached)
	})
}

func TestGetCategoriesDistinctError(t *testing.T) {
	ctx := context.Background()
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("distinct error", func(mt *mtest.T) {
		mr, err := miniredis.Run()
		require.NoError(mt, err)
		defer mr.Close()
		rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
		repo := &productRepository{collection: mt.Coll, redis: rdb}

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(mtest.CommandError{
			Message: "distinct failed",
		}))

		_, err = repo.GetCategories(ctx)
		require.Error(mt, err)
	})
}
