package mongo

import (
	"context"
	"fmt"
	"testing"

	"github.com/kr/pretty"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMongoAction(t *testing.T) {
	mongoInst := &Mongo{}
	err := mongoInst.OpenWithConfig(&MongoConfig{
		Host:       "127.0.0.1",
		Port:       27017,
		DBName:     "test",
		User:       "test",
		Password:   "123456",
		AuthSource: "admin",
	})
	if err != nil {
		fmt.Println("connect", err)
		return
	}
	// err = InsertOne(context.TODO(), mongoInst.DB(), "test", "test", map[string]interface{}{"name": "test"})
	// if err != nil {
	// 	fmt.Println("insert one", err)
	// 	return
	// }

	var result map[string]interface{}
	err = FindOne(context.TODO(), mongoInst.DB(), "test", "test", bson.D{{Key: "name", Value: "test"}}, nil, nil, 0, &result)
	if err != nil {
		fmt.Println("find one", err)
		return
	}
	pretty.Println(result)
}
