// Package service_test kiểm thử logic nghiệp vụ của hệ thống quản lý quyền truy cập
package service_test

import (
	"context"
	"testing"

	acmodels "github.com/aeshield/backend/internal/accesscontrol/models"
	"github.com/aeshield/backend/internal/accesscontrol/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAccessControlRepository là mock cho interface AccessControlRepository
type MockAccessControlRepository struct {
	mock.Mock
}

func (m *MockAccessControlRepository) CreateRule(ctx context.Context, rule *acmodels.AccessRule) error {
	args := m.Called(ctx, rule)
	return args.Error(0)
}

func (m *MockAccessControlRepository) GetRuleByID(ctx context.Context, id interface{}) (*acmodels.AccessRule, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*acmodels.AccessRule), args.Error(1)
}

func (m *MockAccessControlRepository) GetRuleByResource(ctx context.Context, resourceID string, resourceTypes ...string) (*acmodels.AccessRule, error) {
	args := m.Called(ctx, resourceID, resourceTypes)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*acmodels.AccessRule), args.Error(1)
}

func (m *MockAccessControlRepository) UpdateRule(ctx context.Context, id interface{}, rule *acmodels.AccessRule) error {
	args := m.Called(ctx, id, rule)
	return args.Error(0)
}

func (m *MockAccessControlRepository) UpdateRuleByResource(ctx context.Context, resourceID string, rule *acmodels.AccessRule) error {
	args := m.Called(ctx, resourceID, rule)
	return args.Error(0)
}

func (m *MockAccessControlRepository) DeleteRule(ctx context.Context, id interface{}) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAccessControlRepository) DeleteRuleByResource(ctx context.Context, resourceID string) error {
	args := m.Called(ctx, resourceID)
	return args.Error(0)
}

func (m *MockAccessControlRepository) IsOwner(ctx context.Context, resourceID, userID string) (bool, error) {
	args := m.Called(ctx, resourceID, userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockAccessControlRepository) GetWhitelist(ctx context.Context, resourceID string) ([]string, error) {
	args := m.Called(ctx, resourceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func TestCreateRule(t *testing.T) {
	mockRepo := new(MockAccessControlRepository)
	service := service.NewAccessControlService(mockRepo)

	ctx := context.Background()
	resourceID := "test-resource"
	resourceType := "file"
	ownerID := "user-123"
	accessMode := acmodels.AccessModePublic
	whitelist := []string{"user@example.com"}

	mockRepo.On("CreateRule", ctx, mock.AnythingOfType("*models.AccessRule")).Return(nil)

	rule, err := service.CreateRule(ctx, resourceID, resourceType, ownerID, accessMode, whitelist)

	assert.NoError(t, err)
	assert.NotNil(t, rule)
	assert.Equal(t, resourceID, rule.ResourceID)
	assert.Equal(t, resourceType, rule.ResourceType)
	assert.Equal(t, ownerID, rule.OwnerID)
	assert.Equal(t, accessMode, rule.AccessMode)
	assert.Equal(t, whitelist, rule.Whitelist)

	mockRepo.AssertExpectations(t)
}

func TestGetRuleByResource(t *testing.T) {
	mockRepo := new(MockAccessControlRepository)
	service := service.NewAccessControlService(mockRepo)

	ctx := context.Background()
	resourceID := "test-resource"
	rule := &acmodels.AccessRule{
		ResourceID:   resourceID,
		ResourceType: "file",
		OwnerID:      "user-123",
		AccessMode:   acmodels.AccessModePrivate,
		Whitelist:    []string{},
	}

	mockRepo.On("GetRuleByResource", ctx, resourceID, []string(nil)).Return(rule, nil)

	result, err := service.GetRuleByResource(ctx, resourceID)

	assert.NoError(t, err)
	assert.Equal(t, rule, result)

	mockRepo.AssertExpectations(t)
}

func TestUpdateRule(t *testing.T) {
	mockRepo := new(MockAccessControlRepository)
	service := service.NewAccessControlService(mockRepo)

	ctx := context.Background()
	resourceID := "test-resource"
	currentRule := &acmodels.AccessRule{
		ResourceID:   resourceID,
		ResourceType: "file",
		OwnerID:      "user-123",
		AccessMode:   acmodels.AccessModePrivate,
		Whitelist:    []string{},
	}

	accessMode := acmodels.AccessModeWhitelist
	whitelist := []string{"user@example.com"}

	mockRepo.On("GetRuleByResource", ctx, resourceID, []string(nil)).Return(currentRule, nil)
	mockRepo.On("UpdateRuleByResource", ctx, resourceID, mock.AnythingOfType("*models.AccessRule")).Return(nil)

	rule, err := service.UpdateRule(ctx, resourceID, accessMode, whitelist)

	assert.NoError(t, err)
	assert.Equal(t, accessMode, rule.AccessMode)
	assert.Equal(t, whitelist, rule.Whitelist)

	mockRepo.AssertExpectations(t)
}

func TestDeleteRule(t *testing.T) {
	mockRepo := new(MockAccessControlRepository)
	service := service.NewAccessControlService(mockRepo)

	ctx := context.Background()
	resourceID := "test-resource"

	mockRepo.On("DeleteRuleByResource", ctx, resourceID).Return(nil)

	err := service.DeleteRule(ctx, resourceID)

	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

func TestCheckAccess_Public(t *testing.T) {
	mockRepo := new(MockAccessControlRepository)
	service := service.NewAccessControlService(mockRepo)

	ctx := context.Background()
	resourceID := "test-resource"
	rule := &acmodels.AccessRule{
		ResourceID:   resourceID,
		ResourceType: "file",
		OwnerID:      "user-123",
		AccessMode:   acmodels.AccessModePublic,
		Whitelist:    []string{},
	}

	request := &acmodels.CheckAccessRequest{
		ResourceID:    resourceID,
		RequesterID:   "other-user",
		RequesterEmail: "other@example.com",
	}

	mockRepo.On("GetRuleByResource", ctx, resourceID, []string(nil)).Return(rule, nil)

	result := service.CheckAccess(ctx, request)

	assert.True(t, result.HasAccess)
	assert.Equal(t, "Public resource", result.Reason)

	mockRepo.AssertExpectations(t)
}

func TestCheckAccess_Private_Owner(t *testing.T) {
	mockRepo := new(MockAccessControlRepository)
	service := service.NewAccessControlService(mockRepo)

	ctx := context.Background()
	resourceID := "test-resource"
	rule := &acmodels.AccessRule{
		ResourceID:   resourceID,
		ResourceType: "file",
		OwnerID:      "user-123",
		AccessMode:   acmodels.AccessModePrivate,
		Whitelist:    []string{},
	}

	request := &acmodels.CheckAccessRequest{
		ResourceID:    resourceID,
		RequesterID:   "user-123",
		RequesterEmail: "",
	}

	mockRepo.On("GetRuleByResource", ctx, resourceID, []string(nil)).Return(rule, nil)
	mockRepo.On("IsOwner", ctx, resourceID, "user-123").Return(true, nil)

	result := service.CheckAccess(ctx, request)

	assert.True(t, result.HasAccess)
	assert.Equal(t, "Owner access", result.Reason)

	mockRepo.AssertExpectations(t)
}

func TestCheckAccess_Private_NonOwner(t *testing.T) {
	mockRepo := new(MockAccessControlRepository)
	service := service.NewAccessControlService(mockRepo)

	ctx := context.Background()
	resourceID := "test-resource"
	rule := &acmodels.AccessRule{
		ResourceID:   resourceID,
		ResourceType: "file",
		OwnerID:      "user-123",
		AccessMode:   acmodels.AccessModePrivate,
		Whitelist:    []string{},
	}

	request := &acmodels.CheckAccessRequest{
		ResourceID:    resourceID,
		RequesterID:   "other-user",
		RequesterEmail: "",
	}

	mockRepo.On("GetRuleByResource", ctx, resourceID, []string(nil)).Return(rule, nil)
	mockRepo.On("IsOwner", ctx, resourceID, "other-user").Return(false, nil)

	result := service.CheckAccess(ctx, request)

	assert.False(t, result.HasAccess)
	assert.Equal(t, "Access denied: private resource, not owner", result.Reason)

	mockRepo.AssertExpectations(t)
}

func TestCheckAccess_Whitelist_UserID(t *testing.T) {
	mockRepo := new(MockAccessControlRepository)
	service := service.NewAccessControlService(mockRepo)

	ctx := context.Background()
	resourceID := "test-resource"
	rule := &acmodels.AccessRule{
		ResourceID:   resourceID,
		ResourceType: "file",
		OwnerID:      "user-123",
		AccessMode:   acmodels.AccessModeWhitelist,
		Whitelist:    []string{"allowed-user", "another-user"},
	}

	request := &acmodels.CheckAccessRequest{
		ResourceID:    resourceID,
		RequesterID:   "allowed-user",
		RequesterEmail: "",
	}

	mockRepo.On("GetRuleByResource", ctx, resourceID, []string(nil)).Return(rule, nil)

	result := service.CheckAccess(ctx, request)

	assert.True(t, result.HasAccess)
	assert.Equal(t, "User in whitelist", result.Reason)

	mockRepo.AssertExpectations(t)
}

func TestCheckAccess_Whitelist_Email(t *testing.T) {
	mockRepo := new(MockAccessControlRepository)
	service := service.NewAccessControlService(mockRepo)

	ctx := context.Background()
	resourceID := "test-resource"
	rule := &acmodels.AccessRule{
		ResourceID:   resourceID,
		ResourceType: "file",
		OwnerID:      "user-123",
		AccessMode:   acmodels.AccessModeWhitelist,
		Whitelist:    []string{"allowed@example.com", "another@example.com"},
	}

	request := &acmodels.CheckAccessRequest{
		ResourceID:    resourceID,
		RequesterID:   "",
		RequesterEmail: "allowed@example.com",
	}

	mockRepo.On("GetRuleByResource", ctx, resourceID, []string(nil)).Return(rule, nil)

	result := service.CheckAccess(ctx, request)

	assert.True(t, result.HasAccess)
	assert.Equal(t, "Email in whitelist", result.Reason)

	mockRepo.AssertExpectations(t)
}

func TestIsOwner(t *testing.T) {
	mockRepo := new(MockAccessControlRepository)
	service := service.NewAccessControlService(mockRepo)

	ctx := context.Background()
	resourceID := "test-resource"
	userID := "user-123"

	mockRepo.On("IsOwner", ctx, resourceID, userID).Return(true, nil)

	isOwner, err := service.IsOwner(ctx, resourceID, userID)

	assert.NoError(t, err)
	assert.True(t, isOwner)

	mockRepo.AssertExpectations(t)
}

func TestAddToWhitelist(t *testing.T) {
	mockRepo := new(MockAccessControlRepository)
	service := service.NewAccessControlService(mockRepo)

	ctx := context.Background()
	resourceID := "test-resource"
	rule := &acmodels.AccessRule{
		ResourceID:   resourceID,
		ResourceType: "file",
		OwnerID:      "user-123",
		AccessMode:   acmodels.AccessModeWhitelist,
		Whitelist:    []string{"existing@example.com"},
	}

	mockRepo.On("GetRuleByResource", ctx, resourceID, []string(nil)).Return(rule, nil)
	mockRepo.On("UpdateRuleByResource", ctx, resourceID, mock.AnythingOfType("*models.AccessRule")).Return(nil)

	err := service.AddToWhitelist(ctx, resourceID, "new@example.com")

	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

func TestRemoveFromWhitelist(t *testing.T) {
	mockRepo := new(MockAccessControlRepository)
	service := service.NewAccessControlService(mockRepo)

	ctx := context.Background()
	resourceID := "test-resource"
	rule := &acmodels.AccessRule{
		ResourceID:   resourceID,
		ResourceType: "file",
		OwnerID:      "user-123",
		AccessMode:   acmodels.AccessModeWhitelist,
		Whitelist:    []string{"existing@example.com", "toremove@example.com"},
	}

	mockRepo.On("GetRuleByResource", ctx, resourceID, []string(nil)).Return(rule, nil)
	mockRepo.On("UpdateRuleByResource", ctx, resourceID, mock.AnythingOfType("*models.AccessRule")).Return(nil)

	err := service.RemoveFromWhitelist(ctx, resourceID, "toremove@example.com")

	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
}