package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func FindById(ctx context.Context, db *mongo.Client, dbName, collectionName string, id primitive.ObjectID, result any) error {
	return FindOne(ctx, db, dbName, collectionName, filterById(id), nil, nil, 0, result)
}

func FindByIds(ctx context.Context, db *mongo.Client, dbName, collectionName string, ids []primitive.ObjectID, results any) error {
	filter := bson.M{"_id": bson.M{"$in": ids}}
	return FindMany(ctx, db, dbName, collectionName, filter, nil, nil, 0, 0, results)
}

func UpdateById(ctx context.Context, db *mongo.Client, dbName, collectionName string, id primitive.ObjectID, update any) error {
	_, err := UpdateOne(ctx, db, dbName, collectionName, filterById(id), update)
	return err
}

func UpsertById(ctx context.Context, db *mongo.Client, dbName, collectionName string, id primitive.ObjectID, replacement any) error {
	return UpsertOne(ctx, db, dbName, collectionName, filterById(id), replacement)
}

func DeleteById(ctx context.Context, db *mongo.Client, dbName, collectionName string, id primitive.ObjectID) error {
	return DeleteOne(ctx, db, dbName, collectionName, filterById(id))
}

func filterById(id primitive.ObjectID) bson.M {
	return bson.M{"_id": id}
}
