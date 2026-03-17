package storage

import (
	"context"
	"time"

	"github.com/aeshield/backend/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserStorageRepository struct {
	collection *mongo.Collection
}

func NewUserStorageRepository(db *mongo.Database) *UserStorageRepository {
	return &UserStorageRepository{
		collection: db.Collection("user_storage"),
	}
}

func (r *UserStorageRepository) GetByUserID(ctx context.Context, userID string) (*models.UserStorage, error) {
	now := time.Now()
	filter := bson.M{"user_id": userID}
	update := bson.M{
		"$setOnInsert": bson.M{
			"user_id":     userID,
			"used_bytes":  int64(0),
			"file_count":  int64(0),
			"quota_bytes": models.DefaultUserQuotaBytes,
			"updated_at":  now,
		},
	}

	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)

	var storage models.UserStorage
	if err := r.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&storage); err != nil {
		return nil, err
	}

	return &storage, nil
}

func (r *UserStorageRepository) AdjustUsage(ctx context.Context, userID string, usedBytesDelta, fileCountDelta int64) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"user_id": userID},
		bson.M{
			"$setOnInsert": bson.M{
				"user_id":     userID,
				"quota_bytes": models.DefaultUserQuotaBytes,
			},
			"$inc": bson.M{
				"used_bytes": usedBytesDelta,
				"file_count": fileCountDelta,
			},
			"$set": bson.M{"updated_at": time.Now()},
		},
		options.Update().SetUpsert(true),
	)
	return err
}

func (r *UserStorageRepository) SetUsageIfEmpty(ctx context.Context, userID string, usedBytes, fileCount int64) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{
			"user_id":    userID,
			"used_bytes": 0,
			"file_count": 0,
		},
		bson.M{
			"$set": bson.M{
				"used_bytes": usedBytes,
				"file_count": fileCount,
				"updated_at": time.Now(),
			},
		},
	)
	return err
}

func (r *UserStorageRepository) CreateIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "user_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	return err
}
