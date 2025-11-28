package repository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestGetProductsCount_Error(t *testing.T) {
	ctx := context.Background()
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("count error", func(mt *mtest.T) {
		repo := &productRepository{collection: mt.Coll}
		mt.AddMockResponses(mtest.CreateCommandErrorResponse(mtest.CommandError{Message: "count failed"}))

		_, err := repo.GetProductsCount(ctx)
		require.Error(mt, err)
	})
}

func TestGetProductsCountByCategory_Error(t *testing.T) {
	ctx := context.Background()
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("count error", func(mt *mtest.T) {
		repo := &productRepository{collection: mt.Coll}
		mt.AddMockResponses(mtest.CreateCommandErrorResponse(mtest.CommandError{Message: "count failed"}))

		_, err := repo.GetProductsCountByCategory(ctx, "Dogs")
		require.Error(mt, err)
	})
}

func TestGetProductQuantity_NotFound(t *testing.T) {
	ctx := context.Background()
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("no documents", func(mt *mtest.T) {
		repo := &productRepository{collection: mt.Coll}
		mt.AddMockResponses(mtest.CreateCommandErrorResponse(mtest.CommandError{Message: mongo.ErrNoDocuments.Error()}))

		_, err := repo.GetProductQuantity(ctx, "507f1f77bcf86cd799439011")
		require.Error(mt, err)
	})
}
