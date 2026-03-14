package database

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

var (
	ErrUserNotFound = errors.New("user not found")
)

type UserRepository struct {
	collection *mongo.Collection
}

func NewUserRepository(db *MongoDB) *UserRepository {
	return &UserRepository{
		collection: db.Collection("users"),
	}
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		return err
	}

	user.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*models.User, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrUserNotFound
	}

	var user models.User
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) FindByProvider(ctx context.Context, provider, providerID string) (*models.User, error) {
	var user models.User
	err := r.collection.FindOne(ctx, bson.M{
		"providers": bson.M{
			"$elemMatch": bson.M{
				"provider":    provider,
				"provider_id": providerID,
			},
		},
	}).Decode(&user)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	user.UpdatedAt = time.Now()

	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": user.ID},
		bson.M{"$set": bson.M{
			"email":      user.Email,
			"name":       user.Name,
			"avatar":     user.Avatar,
			"providers":  user.Providers,
			"updated_at": user.UpdatedAt,
		}},
	)
	return err
}

func (r *UserRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	return err
}

func (r *UserRepository) MigrateOldSchema(ctx context.Context) error {
	// Tìm các user có schema cũ (có field provider trực tiếp, chưa có providers array)
	cursor, err := r.collection.Find(ctx, bson.M{
		"provider":  bson.M{"$exists": true},
		"providers": bson.M{"$exists": false},
	})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	type OldUser struct {
		ID         primitive.ObjectID `bson:"_id"`
		Provider   string             `bson:"provider"`
		ProviderID string             `bson:"provider_id"`
	}

	for cursor.Next(ctx) {
		var old OldUser
		if err := cursor.Decode(&old); err != nil {
			continue
		}

		providers := []models.LinkedProvider{
			{Provider: old.Provider, ProviderID: old.ProviderID},
		}

		_, _ = r.collection.UpdateOne(ctx,
			bson.M{"_id": old.ID},
			bson.M{
				"$set":   bson.M{"providers": providers},
				"$unset": bson.M{"provider": "", "provider_id": ""},
			},
		)
	}

	return nil
}

func (r *UserRepository) CreateIndexes(ctx context.Context) error {
	// Drop các index cũ từ schema trước (provider + provider_id trực tiếp)
	oldIndexes := []string{
		"provider_1_provider_id_1",
		"email_1",
	}
	for _, name := range oldIndexes {
		// Bỏ qua lỗi nếu index không tồn tại
		_, _ = r.collection.Indexes().DropOne(ctx, name)
	}

	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "providers.provider", Value: 1},
				{Key: "providers.provider_id", Value: 1},
			},
			Options: options.Index().SetUnique(true).SetSparse(true),
		},
		{
			Keys:    bson.D{{Key: "email", Value: 1}},
			Options: options.Index().SetUnique(true).SetSparse(true),
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	return err
}
