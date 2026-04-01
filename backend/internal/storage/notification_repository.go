package storage

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aeshield/backend/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type NotificationListResult struct {
	Items      []*models.Notification
	HasMore    bool
	NextCursor string
}

type NotificationRepository struct {
	collection *mongo.Collection
}

func NewNotificationRepository(db *mongo.Database) *NotificationRepository {
	return &NotificationRepository{
		collection: db.Collection("notifications"),
	}
}

func (r *NotificationRepository) CreateMany(ctx context.Context, notifications []*models.Notification) error {
	if len(notifications) == 0 {
		return nil
	}

	now := time.Now().UTC()
	docs := make([]interface{}, 0, len(notifications))
	for _, notification := range notifications {
		if notification == nil {
			continue
		}
		if notification.CreatedAt.IsZero() {
			notification.CreatedAt = now
		}
		docs = append(docs, notification)
	}

	if len(docs) == 0 {
		return nil
	}

	result, err := r.collection.InsertMany(ctx, docs)
	if err != nil {
		return err
	}

	for index, insertedID := range result.InsertedIDs {
		objectID, ok := insertedID.(primitive.ObjectID)
		if !ok {
			continue
		}
		notifications[index].ID = objectID
	}

	return nil
}

func (r *NotificationRepository) ListByRecipient(ctx context.Context, recipientUserID string, limit int64, cursor string) (*NotificationListResult, error) {
	if limit <= 0 {
		limit = 5
	}
	filter := bson.M{"recipient_user_id": recipientUserID}

	if cursor != "" {
		createdAt, objectID, err := parseNotificationCursor(cursor)
		if err != nil {
			return nil, err
		}
		filter["$or"] = bson.A{
			bson.M{"created_at": bson.M{"$lt": createdAt}},
			bson.M{
				"created_at": createdAt,
				"_id":        bson.M{"$lt": objectID},
			},
		}
	}

	findOptions := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}, {Key: "_id", Value: -1}}).
		SetLimit(limit + 1)

	cursorResult, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursorResult.Close(ctx)

	var notifications []*models.Notification
	if err := cursorResult.All(ctx, &notifications); err != nil {
		return nil, err
	}

	result := &NotificationListResult{}
	if int64(len(notifications)) > limit {
		result.HasMore = true
		notifications = notifications[:limit]
	}
	result.Items = notifications

	if result.HasMore && len(notifications) > 0 {
		last := notifications[len(notifications)-1]
		result.NextCursor = buildNotificationCursor(last.CreatedAt, last.ID)
	}

	return result, nil
}

func (r *NotificationRepository) CountUnreadByRecipient(ctx context.Context, recipientUserID string) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{
		"recipient_user_id": recipientUserID,
		"read_at":            nil,
	})
}

func (r *NotificationRepository) MarkAllRead(ctx context.Context, recipientUserID string, readAt time.Time) (int64, error) {
	result, err := r.collection.UpdateMany(ctx, bson.M{
		"recipient_user_id": recipientUserID,
		"read_at":            nil,
	}, bson.M{
		"$set": bson.M{"read_at": readAt},
	})
	if err != nil {
		return 0, err
	}
	return result.ModifiedCount, nil
}

func (r *NotificationRepository) CreateIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "recipient_user_id", Value: 1}, {Key: "created_at", Value: -1}, {Key: "_id", Value: -1}},
		},
		{
			Keys: bson.D{{Key: "recipient_user_id", Value: 1}, {Key: "read_at", Value: 1}},
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	return err
}

func buildNotificationCursor(createdAt time.Time, id primitive.ObjectID) string {
	return fmt.Sprintf("%s|%s", createdAt.UTC().Format(time.RFC3339Nano), id.Hex())
}

func parseNotificationCursor(cursor string) (time.Time, primitive.ObjectID, error) {
	parts := strings.Split(cursor, "|")
	if len(parts) != 2 {
		return time.Time{}, primitive.NilObjectID, fmt.Errorf("invalid cursor")
	}

	createdAt, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return time.Time{}, primitive.NilObjectID, fmt.Errorf("invalid cursor")
	}

	objectID, err := primitive.ObjectIDFromHex(parts[1])
	if err != nil {
		return time.Time{}, primitive.NilObjectID, fmt.Errorf("invalid cursor")
	}

	return createdAt, objectID, nil
}
