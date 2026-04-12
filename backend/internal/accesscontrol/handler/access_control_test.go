// Package handler_test kiểm thử các endpoint HTTP cho hệ thống quản lý quyền truy cập
package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	acmodels "github.com/aeshield/backend/internal/accesscontrol/models"
	"github.com/aeshield/backend/internal/accesscontrol/service"
	"github.com/aeshield/backend/internal/accesscontrol/handler"
	bmodels "github.com/aeshield/backend/models"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAccessControlRepository là mock cho interface AccessControlRepository
type MockAccessControlRepository struct {
	mock.Mock
}

func normalizeResourceTypes(resourceTypes []string) []string {
	if resourceTypes == nil {
		return []string{}
	}
	return resourceTypes
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
	args := m.Called(ctx, resourceID, normalizeResourceTypes(resourceTypes))
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

// TestApp là cấu trúc để kiểm thử handler
type TestApp struct {
	app  *fiber.App
	repo *MockAccessControlRepository
}

// NewTestApp tạo mới một ứng dụng test
func NewTestApp() *TestApp {
	repo := new(MockAccessControlRepository)
	service := service.NewAccessControlService(repo)
	handler := handler.NewHandler(service)

	app := fiber.New()

	// Middleware để thêm user vào context nếu có
	app.Use(func(c *fiber.Ctx) error {
		// Kiểm tra header Authorization hoặc query param để xác định user
		// Trong thực tế, bạn sẽ có middleware JWT thực sự
		// Nhưng cho mục đích test, chúng ta sẽ giả lập

		// Nếu có header X-Test-User-ID, thì thêm user vào context
		userID := c.Get("X-Test-User-ID")
		if userID != "" {
			email := c.Get("X-Test-User-Email", "test@example.com")
			c.Locals("user", &bmodels.Claims{
				UserID: userID,
				Email:  email,
			})
		}

		return c.Next()
	})

	// Đăng ký các route để kiểm thử
	app.Post("/access/rules", handler.CreateRule)
	app.Get("/access/rules/:resource_id", handler.GetRule)
	app.Patch("/access/rules/:resource_id", handler.UpdateRule)
	app.Delete("/access/rules/:resource_id", handler.DeleteRule)
	app.Post("/access/check", handler.CheckAccess)
	app.Post("/access/whitelist", handler.AddToWhitelist)
	app.Delete("/access/whitelist", handler.RemoveFromWhitelist)

	return &TestApp{
		app:  app,
		repo: repo,
	}
}

func TestCreateRule(t *testing.T) {
	t.Run("Success with valid request", func(t *testing.T) {
		testApp := NewTestApp()

		testApp.repo.On("CreateRule", mock.Anything, mock.AnythingOfType("*models.AccessRule")).Return(nil)

		reqBody := `{"resource_id": "test-resource", "resource_type": "file", "access_mode": "public", "whitelist": ["user@example.com"]}`
		req := httptest.NewRequest("POST", "/access/rules", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Test-User-ID", "user-123")

		resp, err := testApp.app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 201, resp.StatusCode)

		var responseRule acmodels.AccessRule
		json.NewDecoder(resp.Body).Decode(&responseRule)
		assert.Equal(t, "test-resource", responseRule.ResourceID)
		assert.Equal(t, acmodels.AccessModePublic, responseRule.AccessMode)

		testApp.repo.AssertExpectations(t)
	})

	t.Run("Unauthorized without authentication", func(t *testing.T) {
		testApp := NewTestApp()

		reqBody := `{"resource_id": "test-resource", "resource_type": "file", "access_mode": "public"}`
		req := httptest.NewRequest("POST", "/access/rules", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		// Không có header X-Test-User-ID

		resp, err := testApp.app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 401, resp.StatusCode)

		testApp.repo.AssertNotCalled(t, "CreateRule")
	})

	t.Run("Invalid access mode", func(t *testing.T) {
		testApp := NewTestApp()

		reqBody := `{"resource_id": "test-resource", "resource_type": "file", "access_mode": "invalid"}`
		req := httptest.NewRequest("POST", "/access/rules", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Test-User-ID", "user-123")

		resp, err := testApp.app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode)

		testApp.repo.AssertNotCalled(t, "CreateRule")
	})
}

func TestGetRule(t *testing.T) {
	t.Run("Success with authenticated owner", func(t *testing.T) {
		testApp := NewTestApp()

		testApp.repo.On("GetRuleByResource", mock.Anything, "test-resource", []string{}).Return(&acmodels.AccessRule{
			ResourceID:   "test-resource",
			ResourceType: "file",
			OwnerID:      "user-123",
			AccessMode:   acmodels.AccessModePrivate,
			Whitelist:    []string{"allowed@example.com"},
		}, nil)
		testApp.repo.On("IsOwner", mock.Anything, "test-resource", "user-123").Return(true, nil)

		req := httptest.NewRequest("GET", "/access/rules/test-resource", nil)
		req.Header.Set("X-Test-User-ID", "user-123")

		resp, err := testApp.app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var responseRule acmodels.AccessRule
		json.NewDecoder(resp.Body).Decode(&responseRule)
		assert.Equal(t, "test-resource", responseRule.ResourceID)
		assert.Equal(t, []string{"allowed@example.com"}, responseRule.Whitelist) // Owner sees full whitelist

		testApp.repo.AssertExpectations(t)
	})

	t.Run("Success with non-owner or unauthenticated", func(t *testing.T) {
		testApp := NewTestApp()

		testApp.repo.On("GetRuleByResource", mock.Anything, "test-resource", []string{}).Return(&acmodels.AccessRule{
			ResourceID:   "test-resource",
			ResourceType: "file",
			OwnerID:      "owner-123",
			AccessMode:   acmodels.AccessModePrivate,
			Whitelist:    []string{"allowed@example.com"},
		}, nil)
		// Không gọi IsOwner vì không có claims

		req := httptest.NewRequest("GET", "/access/rules/test-resource", nil)
		// Không có header X-Test-User-ID

		resp, err := testApp.app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var responseRule acmodels.AccessRule
		json.NewDecoder(resp.Body).Decode(&responseRule)
		assert.Equal(t, "test-resource", responseRule.ResourceID)
		// Unauthenticated users see the full rule (including whitelist)
		// Only authenticated non-owners see empty whitelist
		assert.Equal(t, []string{"allowed@example.com"}, responseRule.Whitelist)

		testApp.repo.AssertExpectations(t)
	})

	t.Run("Rule not found", func(t *testing.T) {
		testApp := NewTestApp()

		testApp.repo.On("GetRuleByResource", mock.Anything, "non-existent", []string{}).Return((*acmodels.AccessRule)(nil), errors.New("access rule not found"))

		req := httptest.NewRequest("GET", "/access/rules/non-existent", nil)

		resp, err := testApp.app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 404, resp.StatusCode)

		testApp.repo.AssertExpectations(t)
	})
}

func TestCheckAccess(t *testing.T) {
	t.Run("Authenticated user access", func(t *testing.T) {
		testApp := NewTestApp()

		testApp.repo.On("GetRuleByResource", mock.Anything, "test-resource", []string{}).Return(&acmodels.AccessRule{
			ResourceID:   "test-resource",
			ResourceType: "file",
			OwnerID:      "owner-123",
			AccessMode:   acmodels.AccessModePublic,
			Whitelist:    []string{},
		}, nil)

		reqBody := `{"resource_id": "test-resource"}`
		req := httptest.NewRequest("POST", "/access/check", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Test-User-ID", "user-123")
		req.Header.Set("X-Test-User-Email", "user@example.com")

		resp, err := testApp.app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var responseResult acmodels.CheckAccessResult
		json.NewDecoder(resp.Body).Decode(&responseResult)
		assert.True(t, responseResult.HasAccess)
		assert.Equal(t, "Public resource", responseResult.Reason)

		testApp.repo.AssertExpectations(t)
	})

	t.Run("Unauthenticated user access", func(t *testing.T) {
		testApp := NewTestApp()

		testApp.repo.On("GetRuleByResource", mock.Anything, "test-resource", []string{}).Return(&acmodels.AccessRule{
			ResourceID:   "test-resource",
			ResourceType: "file",
			OwnerID:      "owner-123",
			AccessMode:   acmodels.AccessModePrivate,
			Whitelist:    []string{},
		}, nil)
		testApp.repo.On("IsOwner", mock.Anything, "test-resource", "").Return(false, nil)

		reqBody := `{"resource_id": "test-resource"}`
		req := httptest.NewRequest("POST", "/access/check", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		// Không có header X-Test-User-ID

		resp, err := testApp.app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var responseResult acmodels.CheckAccessResult
		json.NewDecoder(resp.Body).Decode(&responseResult)
		assert.False(t, responseResult.HasAccess)
		assert.Equal(t, "Access denied: private resource, not owner", responseResult.Reason)

		testApp.repo.AssertExpectations(t)
	})

	t.Run("Invalid request body", func(t *testing.T) {
		testApp := NewTestApp()

		req := httptest.NewRequest("POST", "/access/check", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")

		resp, err := testApp.app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode)

		testApp.repo.AssertNotCalled(t, "CheckAccess")
	})
}

func TestUpdateRule(t *testing.T) {
	t.Run("Success with owner", func(t *testing.T) {
		testApp := NewTestApp()

		currentRule := &acmodels.AccessRule{
			ResourceID:   "test-resource",
			ResourceType: "file",
			OwnerID:      "user-123",
			AccessMode:   acmodels.AccessModePrivate,
			Whitelist:    []string{},
		}

		testApp.repo.On("GetRuleByResource", mock.Anything, "test-resource", []string{}).Return(currentRule, nil)
		testApp.repo.On("IsOwner", mock.Anything, "test-resource", "user-123").Return(true, nil)
		testApp.repo.On("UpdateRuleByResource", mock.Anything, "test-resource", mock.AnythingOfType("*models.AccessRule")).Return(nil)

		reqBody := `{"access_mode": "whitelist", "whitelist": ["new@example.com"]}`
		req := httptest.NewRequest("PATCH", "/access/rules/test-resource", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Test-User-ID", "user-123")

		resp, err := testApp.app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var responseRule acmodels.AccessRule
		json.NewDecoder(resp.Body).Decode(&responseRule)
		assert.Equal(t, acmodels.AccessModeWhitelist, responseRule.AccessMode)
		assert.Equal(t, []string{"new@example.com"}, responseRule.Whitelist)

		testApp.repo.AssertExpectations(t)
	})

	t.Run("Forbidden for non-owner", func(t *testing.T) {
		testApp := NewTestApp()

		testApp.repo.On("IsOwner", mock.Anything, "test-resource", "other-user").Return(false, nil)

		reqBody := `{"access_mode": "whitelist", "whitelist": ["new@example.com"]}`
		req := httptest.NewRequest("PATCH", "/access/rules/test-resource", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Test-User-ID", "other-user")

		resp, err := testApp.app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 403, resp.StatusCode)

		testApp.repo.AssertExpectations(t)
	})
}

func TestDeleteRule(t *testing.T) {
	t.Run("Success with owner", func(t *testing.T) {
		testApp := NewTestApp()

		testApp.repo.On("IsOwner", mock.Anything, "test-resource", "user-123").Return(true, nil)
		testApp.repo.On("DeleteRuleByResource", mock.Anything, "test-resource").Return(nil)

		req := httptest.NewRequest("DELETE", "/access/rules/test-resource", nil)
		req.Header.Set("X-Test-User-ID", "user-123")

		resp, err := testApp.app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var response map[string]string
		json.NewDecoder(resp.Body).Decode(&response)
		assert.Equal(t, "access rule deleted", response["message"])

		testApp.repo.AssertExpectations(t)
	})

	t.Run("Forbidden for non-owner", func(t *testing.T) {
		testApp := NewTestApp()

		testApp.repo.On("IsOwner", mock.Anything, "test-resource", "other-user").Return(false, nil)

		req := httptest.NewRequest("DELETE", "/access/rules/test-resource", nil)
		req.Header.Set("X-Test-User-ID", "other-user")

		resp, err := testApp.app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 403, resp.StatusCode)

		testApp.repo.AssertExpectations(t)
	})
}

func TestAddToWhitelist(t *testing.T) {
	t.Run("Success with owner", func(t *testing.T) {
		testApp := NewTestApp()

		testApp.repo.On("GetRuleByResource", mock.Anything, "test-resource", []string{}).Return(&acmodels.AccessRule{
			ResourceID:   "test-resource",
			ResourceType: "file",
			OwnerID:      "user-123",
			AccessMode:   acmodels.AccessModeWhitelist,
			Whitelist:    []string{"existing@example.com"},
		}, nil).Once()
		testApp.repo.On("IsOwner", mock.Anything, "test-resource", "user-123").Return(true, nil).Once()
		testApp.repo.On("UpdateRuleByResource", mock.Anything, "test-resource", mock.AnythingOfType("*models.AccessRule")).Return(nil).Once()
		testApp.repo.On("GetRuleByResource", mock.Anything, "test-resource", []string{}).Return(&acmodels.AccessRule{
			ResourceID:   "test-resource",
			ResourceType: "file",
			OwnerID:      "user-123",
			AccessMode:   acmodels.AccessModeWhitelist,
			Whitelist:    []string{"existing@example.com", "new@example.com"},
		}, nil).Once()

		reqBody := `{"resource_id": "test-resource", "user_identifier": "new@example.com"}`
		req := httptest.NewRequest("POST", "/access/whitelist", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Test-User-ID", "user-123")

		resp, err := testApp.app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var responseRule acmodels.AccessRule
		json.NewDecoder(resp.Body).Decode(&responseRule)
		assert.Equal(t, "test-resource", responseRule.ResourceID)
		assert.Contains(t, responseRule.Whitelist, "new@example.com")

		testApp.repo.AssertExpectations(t)
	})

	t.Run("Forbidden for non-owner", func(t *testing.T) {
		testApp := NewTestApp()

		testApp.repo.On("IsOwner", mock.Anything, "test-resource", "other-user").Return(false, nil)

		reqBody := `{"resource_id": "test-resource", "user_identifier": "new@example.com"}`
		req := httptest.NewRequest("POST", "/access/whitelist", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Test-User-ID", "other-user")

		resp, err := testApp.app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 403, resp.StatusCode)

		testApp.repo.AssertExpectations(t)
	})
}

func TestRemoveFromWhitelist(t *testing.T) {
	t.Run("Success with owner", func(t *testing.T) {
		testApp := NewTestApp()

		testApp.repo.On("GetRuleByResource", mock.Anything, "test-resource", []string{}).Return(&acmodels.AccessRule{
			ResourceID:   "test-resource",
			ResourceType: "file",
			OwnerID:      "user-123",
			AccessMode:   acmodels.AccessModeWhitelist,
			Whitelist:    []string{"remaining@example.com", "toremove@example.com"},
		}, nil)
		testApp.repo.On("IsOwner", mock.Anything, "test-resource", "user-123").Return(true, nil)
		testApp.repo.On("UpdateRuleByResource", mock.Anything, "test-resource", mock.AnythingOfType("*models.AccessRule")).Return(nil).Once()
		testApp.repo.On("GetRuleByResource", mock.Anything, "test-resource", []string{}).Return(&acmodels.AccessRule{
			ResourceID:   "test-resource",
			ResourceType: "file",
			OwnerID:      "user-123",
			AccessMode:   acmodels.AccessModeWhitelist,
			Whitelist:    []string{"remaining@example.com"},
		}, nil)

		reqBody := `{"resource_id": "test-resource", "user_identifier": "toremove@example.com"}`
		req := httptest.NewRequest("DELETE", "/access/whitelist", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Test-User-ID", "user-123")

		resp, err := testApp.app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var responseRule acmodels.AccessRule
		json.NewDecoder(resp.Body).Decode(&responseRule)
		assert.Equal(t, "test-resource", responseRule.ResourceID)
		assert.NotContains(t, responseRule.Whitelist, "toremove@example.com")

		testApp.repo.AssertExpectations(t)
	})
}