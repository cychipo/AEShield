// Package repository chứa triển khai thực tế cho interface AccessControlRepository
package repository

import (
	"context"
	"fmt"

	"github.com/aeshield/backend/internal/accesscontrol/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// MongoAccessControlRepository là triển khai cụ thể sử dụng MongoDB
type MongoAccessControlRepository struct {
	Collection *mongo.Collection
}

// NewMongoAccessControlRepository tạo mới một instance của MongoAccessControlRepository
func NewMongoAccessControlRepository(collection *mongo.Collection) *MongoAccessControlRepository {
	return &MongoAccessControlRepository{
		Collection: collection,
	}
}

// CreateRule tạo mới một quy tắc truy cập
func (r *MongoAccessControlRepository) CreateRule(ctx context.Context, rule *models.AccessRule) error {
	_, err := r.Collection.InsertOne(ctx, rule)
	return err
}

// GetRuleByID lấy quy tắc truy cập theo ID
func (r *MongoAccessControlRepository) GetRuleByID(ctx context.Context, id interface{}) (*models.AccessRule, error) {
	var rule models.AccessRule
	err := r.Collection.FindOne(ctx, bson.M{"_id": id}).Decode(&rule)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("access rule not found")
		}
		return nil, err
	}
	return &rule, nil
}

// GetRuleByResource lấy quy tắc truy cập theo ID tài nguyên
func (r *MongoAccessControlRepository) GetRuleByResource(ctx context.Context, resourceID string, resourceTypes ...string) (*models.AccessRule, error) {
	filter := bson.M{"resource_id": resourceID}
	if len(resourceTypes) > 0 {
		filter["resource_type"] = bson.M{"$in": resourceTypes}
	}

	var rule models.AccessRule
	err := r.Collection.FindOne(ctx, filter).Decode(&rule)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("access rule not found")
		}
		return nil, err
	}
	return &rule, nil
}

// UpdateRule cập nhật một quy tắc truy cập
func (r *MongoAccessControlRepository) UpdateRule(ctx context.Context, id interface{}, rule *models.AccessRule) error {
	rule.ID = id.(primitive.ObjectID) // Convert to ObjectID for internal use
	_, err := r.Collection.ReplaceOne(ctx, bson.M{"_id": id}, rule)
	return err
}

// UpdateRuleByResource cập nhật quy tắc truy cập theo ID tài nguyên
func (r *MongoAccessControlRepository) UpdateRuleByResource(ctx context.Context, resourceID string, rule *models.AccessRule) error {
	filter := bson.M{"resource_id": resourceID}
	update := bson.M{
		"$set": bson.M{
			"access_mode":  rule.AccessMode,
			"whitelist":    rule.Whitelist,
			"updated_at":   rule.UpdatedAt,
		},
	}
	result, err := r.Collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("access rule not found")
	}
	return nil
}

// DeleteRule xóa một quy tắc truy cập
func (r *MongoAccessControlRepository) DeleteRule(ctx context.Context, id interface{}) error {
	_, err := r.Collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

// DeleteRuleByResource xóa quy tắc truy cập theo ID tài nguyên
func (r *MongoAccessControlRepository) DeleteRuleByResource(ctx context.Context, resourceID string) error {
	_, err := r.Collection.DeleteOne(ctx, bson.M{"resource_id": resourceID})
	return err
}

// IsOwner kiểm tra xem người dùng có phải là chủ sở hữu của tài nguyên hay không
func (r *MongoAccessControlRepository) IsOwner(ctx context.Context, resourceID, userID string) (bool, error) {
	filter := bson.M{
		"resource_id": resourceID,
		"owner_id":    userID,
	}

	count, err := r.Collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetWhitelist lấy danh sách trắng của một tài nguyên
func (r *MongoAccessControlRepository) GetWhitelist(ctx context.Context, resourceID string) ([]string, error) {
	var rule models.AccessRule
	err := r.Collection.FindOne(ctx, bson.M{"resource_id": resourceID}).Decode(&rule)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("access rule not found")
		}
		return nil, err
	}

	return rule.Whitelist, nil
}