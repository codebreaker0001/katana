package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func UpdateOne(ctx context.Context, collectionName string, filter bson.M, data interface{}, option *options.FindOneAndUpdateOptions) error {
	collection := link.Collection(collectionName)
	return collection.FindOneAndUpdate(ctx, filter, bson.M{"$set": data}, option).Err()
}

func UpsertOne(ctx context.Context, collectionName string, filter bson.M, data interface{}) error {
	return UpdateOne(ctx, collectionName, filter, data, options.FindOneAndUpdate().SetUpsert(true))
}

func AddAdmin(ctx context.Context, admin AdminUser) error {
	return UpsertOne(ctx, AdminCollection, bson.M{UsernameKey: admin.Username}, admin)
}
