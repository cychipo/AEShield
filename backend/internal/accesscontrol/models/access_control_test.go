// Package models_test kiểm thử các cấu trúc dữ liệu liên quan đến quản lý quyền truy cập
package models_test

import (
	"testing"
	"time"

	"github.com/aeshield/backend/internal/accesscontrol/models"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestAccessRule_Validation(t *testing.T) {
	t.Run("Valid AccessRule", func(t *testing.T) {
		rule := &models.AccessRule{
			ID:           primitive.NewObjectID(),
			ResourceID:   "test-resource",
			ResourceType: "file",
			OwnerID:      "user-123",
			AccessMode:   models.AccessModePublic,
			Whitelist:    []string{"user@example.com"},
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		assert.NotEmpty(t, rule.ID)
		assert.Equal(t, "test-resource", rule.ResourceID)
		assert.Equal(t, "file", rule.ResourceType)
		assert.Equal(t, "user-123", rule.OwnerID)
		assert.Equal(t, models.AccessModePublic, rule.AccessMode)
		assert.Len(t, rule.Whitelist, 1)
	})

	t.Run("AccessRule with empty whitelist", func(t *testing.T) {
		rule := &models.AccessRule{
			ResourceID:   "test-resource",
			ResourceType: "file",
			OwnerID:      "user-123",
			AccessMode:   models.AccessModePrivate,
			Whitelist:    []string{},
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		assert.Empty(t, rule.Whitelist)
		assert.Equal(t, models.AccessModePrivate, rule.AccessMode)
	})
}

func TestAccessMode_Values(t *testing.T) {
	assert.Equal(t, models.AccessMode("public"), models.AccessModePublic)
	assert.Equal(t, models.AccessMode("private"), models.AccessModePrivate)
	assert.Equal(t, models.AccessMode("whitelist"), models.AccessModeWhitelist)
}

func TestCheckAccessRequest(t *testing.T) {
	req := &models.CheckAccessRequest{
		ResourceID:    "test-resource",
		RequesterID:   "user-123",
		RequesterEmail: "user@example.com",
	}

	assert.Equal(t, "test-resource", req.ResourceID)
	assert.Equal(t, "user-123", req.RequesterID)
	assert.Equal(t, "user@example.com", req.RequesterEmail)
}

func TestCheckAccessResult(t *testing.T) {
	result := &models.CheckAccessResult{
		HasAccess: true,
		Reason:    "Test reason",
		Error:     nil,
	}

	assert.True(t, result.HasAccess)
	assert.Equal(t, "Test reason", result.Reason)
	assert.Nil(t, result.Error)
}