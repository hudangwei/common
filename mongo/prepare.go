package mongo

import (
	"context"

	"github.com/hudangwei/common/util/ctxutil"
	"go.mongodb.org/mongo-driver/mongo"
)

func Prepare(ctx context.Context, db *mongo.Client, dbName string, collectionName string) (context.Context, context.CancelFunc, *mongo.Collection) {
	ctx, cancel := ctxutil.Timeout(ctx, 30)
	collection := db.Database(dbName).Collection(collectionName)
	return ctx, cancel, collection
}
