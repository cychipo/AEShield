package storage

import (
	"context"
	"errors"
	"time"

	"github.com/aeshield/backend/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var ErrFileNotFound = errors.New("file not found")

type FileRepository struct {
	collection *mongo.Collection
}

func NewFileRepository(db *mongo.Database) *FileRepository {
	return &FileRepository{
		collection: db.Collection("files"),
	}
}

func (r *FileRepository) Create(ctx context.Context, file *models.FileMetadata) error {
	file.CreatedAt = time.Now()
	file.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, file)
	if err != nil {
		return err
	}

	file.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *FileRepository) FindByID(ctx context.Context, id string) (*models.FileMetadata, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrFileNotFound
	}

	var file models.FileMetadata
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&file)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrFileNotFound
		}
		return nil, err
	}

	return &file, nil
}

func (r *FileRepository) FindByOwner(ctx context.Context, ownerID string) ([]*models.FileMetadata, error) {
	cursor, err := r.collection.Find(ctx,
		bson.M{"owner_id": ownerID},
		options.Find().SetSort(bson.M{"created_at": -1}),
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var files []*models.FileMetadata
	if err := cursor.All(ctx, &files); err != nil {
		return nil, err
	}

	return files, nil
}

func (r *FileRepository) FindSharedWithUser(ctx context.Context, userID string) ([]*models.FileMetadata, error) {
	cursor, err := r.collection.Find(ctx,
		bson.M{
			"access_mode": models.AccessModeWhitelist,
			"whitelist":   userID,
		},
		options.Find().SetSort(bson.M{"created_at": -1}),
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var files []*models.FileMetadata
	if err := cursor.All(ctx, &files); err != nil {
		return nil, err
	}

	return files, nil
}

func (r *FileRepository) FindByPublicCID(ctx context.Context, cid string) (*models.FileMetadata, error) {
	var file models.FileMetadata
	err := r.collection.FindOne(ctx, bson.M{"public_cid": cid}).Decode(&file)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrFileNotFound
		}
		return nil, err
	}
	return &file, nil
}

func (r *FileRepository) Update(ctx context.Context, file *models.FileMetadata) error {
	file.UpdatedAt = time.Now()

	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": file.ID},
		bson.M{"$set": file},
	)
	return err
}

func (r *FileRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return ErrFileNotFound
	}

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return ErrFileNotFound
	}

	return nil
}

func (r *FileRepository) CreateIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "owner_id", Value: 1}},
			Options: options.Index(),
		},
		{
			Keys:    bson.D{{Key: "public_cid", Value: 1}},
			Options: options.Index().SetUnique(true).SetSparse(true),
		},
		{
			Keys:    bson.D{{Key: "access_mode", Value: 1}, {Key: "whitelist", Value: 1}, {Key: "created_at", Value: -1}},
			Options: options.Index(),
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	return err
}
