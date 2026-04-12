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

var ErrJobNotFound = errors.New("job not found")

type JobRepository struct {
	collection *mongo.Collection
}

func NewJobRepository(db *mongo.Database) *JobRepository {
	return &JobRepository{collection: db.Collection("jobs")}
}

func (r *JobRepository) Create(ctx context.Context, job *models.Job) error {
	now := time.Now()
	job.CreatedAt = now
	job.UpdatedAt = now
	result, err := r.collection.InsertOne(ctx, job)
	if err != nil {
		return err
	}
	job.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *JobRepository) FindByID(ctx context.Context, id string) (*models.Job, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrJobNotFound
	}
	var job models.Job
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&job)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrJobNotFound
		}
		return nil, err
	}
	return &job, nil
}

func (r *JobRepository) FindByUserID(ctx context.Context, userID string, status string, limit, offset int64) ([]*models.Job, int64, error) {
	filter := bson.M{"user_id": userID}
	if status != "" {
		filter["status"] = status
	}
	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	opts := options.Find().SetSort(bson.M{"created_at": -1}).SetSkip(offset).SetLimit(limit)
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)
	var jobs []*models.Job
	if err := cursor.All(ctx, &jobs); err != nil {
		return nil, 0, err
	}
	return jobs, count, nil
}

func (r *JobRepository) Update(ctx context.Context, job *models.Job) error {
	job.UpdatedAt = time.Now()
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": job.ID}, bson.M{"$set": job})
	return err
}

func (r *JobRepository) MarkCancelled(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return ErrJobNotFound
	}
	now := time.Now()
	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, bson.M{"$set": bson.M{"status": models.JobStatusCancelled, "updated_at": now, "completed_at": now}})
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return ErrJobNotFound
	}
	return nil
}

func (r *JobRepository) DeleteOlderThan(ctx context.Context, before time.Time) error {
	_, err := r.collection.DeleteMany(ctx, bson.M{"completed_at": bson.M{"$lt": before}})
	return err
}

func (r *JobRepository) CreateIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "created_at", Value: -1}}, Options: options.Index()},
		{Keys: bson.D{{Key: "status", Value: 1}}, Options: options.Index()},
		{Keys: bson.D{{Key: "completed_at", Value: 1}}, Options: options.Index()},
	}
	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	return err
}
