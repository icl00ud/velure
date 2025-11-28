package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestGetProductsByName_Success(t *testing.T) {
	ctx := context.Background()
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("find by name", func(mt *mtest.T) {
		repo := &productRepository{collection: mt.Coll}
		id := primitive.NewObjectID()
		now := time.Now()
		ns := fmt.Sprintf("%s.%s", mt.DB.Name(), mt.Coll.Name())

		mt.AddMockResponses(
			mtest.CreateCursorResponse(1, ns, mtest.FirstBatch, bson.D{
				{Key: "_id", Value: id},
				{Key: "name", Value: "Toy"},
				{Key: "price", Value: 10.0},
				{Key: "quantity", Value: 5},
				{Key: "dimensions", Value: bson.D{}},
				{Key: "images", Value: bson.A{}},
				{Key: "dt_created", Value: now},
				{Key: "dt_updated", Value: now},
			}),
			mtest.CreateCursorResponse(0, ns, mtest.NextBatch),
			mtest.CreateSuccessResponse(),
		)

		products, err := repo.GetProductsByName(ctx, "Toy")
		require.NoError(mt, err)
		require.Len(mt, products, 1)
		require.Equal(mt, "Toy", products[0].Name)
	})
}

func TestGetProductsByName_FindError(t *testing.T) {
	ctx := context.Background()
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("find error", func(mt *mtest.T) {
		repo := &productRepository{collection: mt.Coll}
		mt.AddMockResponses(mtest.CreateCommandErrorResponse(mtest.CommandError{Message: "find fail"}))

		_, err := repo.GetProductsByName(ctx, "Toy")
		require.Error(mt, err)
	})
}
