package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	moptions "go.mongodb.org/mongo-driver/mongo/options"
)

// 插入一条记录。
func InsertOne(ctx context.Context, db *mongo.Client, dbName, collectionName string, document any) error {
	ctx, cancel, coll := Prepare(ctx, db, dbName, collectionName)
	defer cancel()

	if _, err := coll.InsertOne(ctx, document); err != nil {
		return err
	}

	return nil
}

// 插入多条记录。
func InsertMany(ctx context.Context, db *mongo.Client, dbName, collectionName string, documents []any) error {
	ctx, cancel, coll := Prepare(ctx, db, dbName, collectionName)
	defer cancel()

	if _, err := coll.InsertMany(ctx, documents); err != nil {
		return err
	}

	return nil
}

// 查询一条记录。
func FindOne(ctx context.Context, db *mongo.Client, dbName, collectionName string, filter, fields, sort any, cursor int, result any) error {
	ctx, cancel, coll := Prepare(ctx, db, dbName, collectionName)
	defer cancel()

	mopts := moptions.FindOne()
	mopts.SetProjection(fields)
	mopts.SetSort(sort)
	mopts.SetSkip(int64(cursor))

	if err := coll.FindOne(ctx, filter, mopts).Decode(result); err != nil {
		return err
	}

	return nil
}

// 查询多条记录。
func FindMany(ctx context.Context, db *mongo.Client, dbName, collectionName string, filter, fields, sort any, cursor, size int, results any) error {
	ctx, cancel, coll := Prepare(ctx, db, dbName, collectionName)
	defer cancel()

	mopts := moptions.Find()
	mopts.SetProjection(fields)
	mopts.SetSort(sort)
	mopts.SetSkip(int64(cursor))
	mopts.SetLimit(int64(size))

	cur, err := coll.Find(ctx, filter, mopts)
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	if err = cur.All(ctx, results); err != nil {
		return err
	}

	return nil
}

// 查询符合条件的记录数。
func Count(ctx context.Context, db *mongo.Client, dbName, collectionName string, filter any) (int, error) {
	ctx, cancel, coll := Prepare(ctx, db, dbName, collectionName)
	defer cancel()

	count, err := coll.CountDocuments(ctx, filter)
	if err != nil {
		return 0, err
	}

	return int(count), nil
}

// 更新一条记录。
func UpdateOne(ctx context.Context, db *mongo.Client, dbName, collectionName string, filter, update any) (*mongo.UpdateResult, error) {
	ctx, cancel, coll := Prepare(ctx, db, dbName, collectionName)
	defer cancel()

	result, err := coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// 更新多条记录。
func UpdateMany(ctx context.Context, db *mongo.Client, dbName, collectionName string, filter, update any) (*mongo.UpdateResult, error) {
	ctx, cancel, coll := Prepare(ctx, db, dbName, collectionName)
	defer cancel()

	result, err := coll.UpdateMany(ctx, filter, update)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// 更新或插入一条记录。
func UpsertOne(ctx context.Context, db *mongo.Client, dbName, collectionName string, filter, replacement any) error {
	ctx, cancel, coll := Prepare(ctx, db, dbName, collectionName)
	defer cancel()

	mopts := moptions.Replace().SetUpsert(true)
	if _, err := coll.ReplaceOne(ctx, filter, replacement, mopts); err != nil {
		return err
	}

	return nil
}

// 删除一条记录。
// 如果匹配多条记录，则随机删除一条记录。
func DeleteOne(ctx context.Context, db *mongo.Client, dbName, collectionName string, filter any) error {
	ctx, cancel, coll := Prepare(ctx, db, dbName, collectionName)
	defer cancel()

	if _, err := coll.DeleteOne(ctx, filter); err != nil {
		return err
	}

	return nil
}

// 删除多条记录。
func DeleteMany(ctx context.Context, db *mongo.Client, dbName, collectionName string, filter any) error {
	ctx, cancel, coll := Prepare(ctx, db, dbName, collectionName)
	defer cancel()

	if _, err := coll.DeleteMany(ctx, filter); err != nil {
		return err
	}

	return nil
}

// 执行Pipeline查询。
func Pipe(ctx context.Context, db *mongo.Client, dbName, collectionName string, pipeline, result any) error {
	ctx, cancel, coll := Prepare(ctx, db, dbName, collectionName)
	defer cancel()

	cur, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	if err = cur.Decode(result); err != nil {
		return err
	}

	return nil
}
