package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
)

func Drop(ctx context.Context, db *mongo.Client, dbName, collectionName string) error {
	ctx, cancel, coll := Prepare(ctx, db, dbName, collectionName)
	defer cancel()
	return coll.Drop(ctx)
}
