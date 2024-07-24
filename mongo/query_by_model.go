package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
)

type FindModel struct {
	CollectionName string
	Query          any
	Fields         any
	Sort           any
	Cursor         int
	Size           int
	Results        any
}

func FindByModel(ctx context.Context, db *mongo.Client, dbName string, model FindModel) error {
	return FindMany(ctx, db, dbName, model.CollectionName, model.Query, model.Fields, model.Sort, model.Cursor, model.Size, model.Results)
}

func FindOneByModel(ctx context.Context, db *mongo.Client, dbName string, model FindModel) error {
	return FindOne(ctx, db, dbName, model.CollectionName, model.Query, model.Fields, model.Sort, model.Cursor, model.Results)
}
