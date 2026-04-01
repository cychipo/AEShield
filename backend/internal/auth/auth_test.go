package auth

import (
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/aeshield/backend/internal/config"
	"github.com/aeshield/backend/internal/database"
	"github.com/aeshield/backend/models"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/oauth2"
)

type MockUserRepository struct {
	users map[string]*models.User // key: email
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users: make(map[string]*models.User),
	}
}

func (m *MockUserRepository) FindByID(ctx context.Context, id string) (*models.User, error) {
	for _, user := range m.users {
		if user.ID.Hex() == id {
			return user, nil
		}
	}
	return nil, database.ErrUserNotFound
}

func (m *MockUserRepository) FindByProvider(ctx context.Context, provider, providerID string) (*models.User, error) {
	for _, user := range m.users {
		for _, p := range user.Providers {
			if p.Provider == provider && p.ProviderID == providerID {
				return user, nil
			}
		}
	}
	return nil, database.ErrUserNotFound
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	user, ok := m.users[email]
	if !ok {
		return nil, database.ErrUserNotFound
	}
	return user, nil
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	user.ID = primitive.NewObjectID()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	m.users[user.Email] = user
	return nil
}

func (m *MockUserRepository) Update(ctx context.Context, user *models.User) error {
	user.UpdatedAt = time.Now()
	m.users[user.Email] = user
	return nil
}

func setupTestApp() *fiber.App {
	cfg := config.Load()
	mockRepo := NewMockUserRepository()
	service := NewService(cfg, mockRepo)
	handler := NewHandler(service)

	app := fiber.New()

	api := app.Group("/api/v1")
	api.Get("/auth/urls", handler.GetAuthURLs)
	api.Get("/auth/google", handler.GoogleLogin)
	api.Get("/auth/google/callback", handler.GoogleCallback)
	api.Get("/auth/github", handler.GitHubLogin)
	api.Get("/auth/github/callback", handler.GitHubCallback)

	return app
}

func TestGetAuthURLs(t *testing.T) {
	app := setupTestApp()

	req := httptest.NewRequest("GET", "/api/v1/auth/urls", nil)
	resp, err := app.Test(req)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var result map[string]string
	json.Unmarshal(body, &result)

	assert.Contains(t, result, "google")
	assert.Contains(t, result, "github")
	assert.NotEmpty(t, result["google"])
	assert.NotEmpty(t, result["github"])
	assert.Contains(t, result["google"], "accounts.google.com")
	assert.Contains(t, result["github"], "github.com")
}

func TestGoogleLogin(t *testing.T) {
	app := setupTestApp()

	req := httptest.NewRequest("GET", "/api/v1/auth/google", nil)
	resp, err := app.Test(req)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var result map[string]string
	json.Unmarshal(body, &result)

	assert.Contains(t, result, "url")
	assert.Contains(t, result["url"], "accounts.google.com")
	assert.Contains(t, result["url"], "client_id=")
	assert.Contains(t, result["url"], "redirect_uri=")
	assert.Contains(t, result["url"], "response_type=code")
}

func TestGitHubLogin(t *testing.T) {
	app := setupTestApp()

	req := httptest.NewRequest("GET", "/api/v1/auth/github", nil)
	resp, err := app.Test(req)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var result map[string]string
	json.Unmarshal(body, &result)

	assert.Contains(t, result, "url")
	assert.Contains(t, result["url"], "github.com")
	assert.Contains(t, result["url"], "client_id=")
	assert.Contains(t, result["url"], "redirect_uri=")
}

func TestGoogleCallback_MissingCode(t *testing.T) {
	app := setupTestApp()

	req := httptest.NewRequest("GET", "/api/v1/auth/google/callback", nil)
	resp, err := app.Test(req)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var result map[string]string
	json.Unmarshal(body, &result)
	assert.Equal(t, "missing code", result["error"])
}

func TestGitHubCallback_MissingCode(t *testing.T) {
	app := setupTestApp()

	req := httptest.NewRequest("GET", "/api/v1/auth/github/callback", nil)
	resp, err := app.Test(req)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var result map[string]string
	json.Unmarshal(body, &result)
	assert.Equal(t, "missing code", result["error"])
}

func TestMe_Unauthorized(t *testing.T) {
	cfg := config.Load()
	mockRepo := NewMockUserRepository()
	service := NewService(cfg, mockRepo)
	handler := NewHandler(service)

	app := fiber.New()
	api := app.Group("/api/v1")
	api.Get("/auth/me", JWTMiddleware(cfg.JWTSecret), handler.Me)

	req := httptest.NewRequest("GET", "/api/v1/auth/me", nil)
	resp, err := app.Test(req)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestMe_ValidToken(t *testing.T) {
	cfg := config.Load()
	cfg.JWTSecret = "test-secret"
	mockRepo := NewMockUserRepository()
	service := NewService(cfg, mockRepo)
	handler := NewHandler(service)

	app := fiber.New()
	api := app.Group("/api/v1")
	api.Get("/auth/me", JWTMiddleware(cfg.JWTSecret), handler.Me)

	claims := &models.Claims{
		UserID:   "test-user-123",
		Email:    "test@example.com",
		Provider: "google",
		Name:     "Test User",
	}
	tokenString, err := service.GenerateJWT(claims)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/api/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	resp, err := app.Test(req)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestLookupUserByEmail_ValidToken(t *testing.T) {
	cfg := config.Load()
	cfg.JWTSecret = "test-secret"
	mockRepo := NewMockUserRepository()
	mockRepo.users["friend@example.com"] = &models.User{
		ID:     primitive.NewObjectID(),
		Email:  "friend@example.com",
		Name:   "Friend User",
		Avatar: "https://example.com/avatar.png",
	}
	service := NewService(cfg, mockRepo)
	handler := NewHandler(service)

	app := fiber.New()
	api := app.Group("/api/v1")
	api.Get("/users/lookup", JWTMiddleware(cfg.JWTSecret), handler.LookupUserByEmail)

	claims := &models.Claims{UserID: "owner-1", Email: "owner@example.com"}
	tokenString, err := service.GenerateJWT(claims)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/api/v1/users/lookup?email=friend@example.com", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var result UserLookupResult
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.Equal(t, "friend@example.com", result.Email)
	assert.Equal(t, "Friend User", result.Name)
	assert.NotEmpty(t, result.ID)
}

func TestLookupUserByEmail_NotFound(t *testing.T) {
	cfg := config.Load()
	cfg.JWTSecret = "test-secret"
	mockRepo := NewMockUserRepository()
	service := NewService(cfg, mockRepo)
	handler := NewHandler(service)

	app := fiber.New()
	api := app.Group("/api/v1")
	api.Get("/users/lookup", JWTMiddleware(cfg.JWTSecret), handler.LookupUserByEmail)

	claims := &models.Claims{UserID: "owner-1", Email: "owner@example.com"}
	tokenString, err := service.GenerateJWT(claims)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/api/v1/users/lookup?email=missing@example.com", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

func TestLookupUserByEmail_RequireEmail(t *testing.T) {
	cfg := config.Load()
	cfg.JWTSecret = "test-secret"
	mockRepo := NewMockUserRepository()
	service := NewService(cfg, mockRepo)
	handler := NewHandler(service)

	app := fiber.New()
	api := app.Group("/api/v1")
	api.Get("/users/lookup", JWTMiddleware(cfg.JWTSecret), handler.LookupUserByEmail)

	claims := &models.Claims{UserID: "owner-1", Email: "owner@example.com"}
	tokenString, err := service.GenerateJWT(claims)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/api/v1/users/lookup", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestResolveUsersByID_ValidToken(t *testing.T) {
	cfg := config.Load()
	cfg.JWTSecret = "test-secret"
	mockRepo := NewMockUserRepository()
	user := &models.User{
		ID:     primitive.NewObjectID(),
		Email:  "friend@example.com",
		Name:   "Friend User",
		Avatar: "https://example.com/avatar.png",
	}
	mockRepo.users[user.Email] = user
	service := NewService(cfg, mockRepo)
	handler := NewHandler(service)

	app := fiber.New()
	api := app.Group("/api/v1")
	api.Post("/users/resolve", JWTMiddleware(cfg.JWTSecret), handler.ResolveUsersByID)

	claims := &models.Claims{UserID: "owner-1", Email: "owner@example.com"}
	tokenString, err := service.GenerateJWT(claims)
	require.NoError(t, err)

	body := strings.NewReader(`{"ids":["` + user.ID.Hex() + `","missing-id"]}`)
	req := httptest.NewRequest("POST", "/api/v1/users/resolve", body)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var result []UserLookupResult
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	require.Len(t, result, 1)
	assert.Equal(t, user.ID.Hex(), result[0].ID)
	assert.Equal(t, user.Email, result[0].Email)
}

func TestFindOrCreateUser_CreateNew(t *testing.T) {
	cfg := config.Load()
	mockRepo := NewMockUserRepository()
	service := NewService(cfg, mockRepo)

	ctx := context.Background()
	user, err := service.FindOrCreateUser(ctx, "google", "google-123", "test@example.com", "Test User", "https://example.com/avatar.jpg")

	require.NoError(t, err)
	assert.NotNil(t, user)
	assert.NotEqual(t, primitive.ObjectID{}, user.ID)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "Test User", user.Name)
	assert.Equal(t, "https://example.com/avatar.jpg", user.Avatar)
	assert.Len(t, user.Providers, 1)
	assert.Equal(t, "google", user.Providers[0].Provider)
	assert.Equal(t, "google-123", user.Providers[0].ProviderID)
	assert.False(t, user.CreatedAt.IsZero())
}

func TestFindOrCreateUser_UpdateExisting(t *testing.T) {
	cfg := config.Load()
	mockRepo := NewMockUserRepository()

	existingUser := &models.User{
		Email:  "test@example.com",
		Name:   "Old Name",
		Avatar: "https://example.com/old.jpg",
		Providers: []models.LinkedProvider{
			{Provider: "google", ProviderID: "google-123"},
		},
	}
	mockRepo.Create(context.Background(), existingUser)

	service := NewService(cfg, mockRepo)

	user, err := service.FindOrCreateUser(context.Background(), "google", "google-123", "test@example.com", "New Name", "https://example.com/new.jpg")

	require.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "New Name", user.Name)
	assert.Equal(t, "https://example.com/new.jpg", user.Avatar)
}

func TestFindOrCreateUser_MergeByEmail(t *testing.T) {
	cfg := config.Load()
	mockRepo := NewMockUserRepository()

	// Tạo user đã tồn tại với Google
	existingUser := &models.User{
		Email:  "same@example.com",
		Name:   "Same User",
		Avatar: "https://example.com/avatar.jpg",
		Providers: []models.LinkedProvider{
			{Provider: "google", ProviderID: "google-111"},
		},
	}
	mockRepo.Create(context.Background(), existingUser)

	service := NewService(cfg, mockRepo)

	// Login bằng GitHub với cùng email
	user, err := service.FindOrCreateUser(context.Background(), "github", "github-222", "same@example.com", "Same User", "https://example.com/avatar.jpg")

	require.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "same@example.com", user.Email)
	// Phải có 2 providers
	assert.Len(t, user.Providers, 2)
}

func TestFindOrCreateUser_NoChange(t *testing.T) {
	cfg := config.Load()
	mockRepo := NewMockUserRepository()

	existingUser := &models.User{
		Email:  "same@example.com",
		Name:   "Same Name",
		Avatar: "https://example.com/same.jpg",
		Providers: []models.LinkedProvider{
			{Provider: "github", ProviderID: "github-456"},
		},
	}
	mockRepo.Create(context.Background(), existingUser)

	service := NewService(cfg, mockRepo)

	user, err := service.FindOrCreateUser(context.Background(), "github", "github-456", "same@example.com", "Same Name", "https://example.com/same.jpg")

	require.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "same@example.com", user.Email)
	assert.Equal(t, "Same Name", user.Name)
}

func TestGenerateJWT(t *testing.T) {
	cfg := config.Load()
	service := NewService(cfg, nil)

	claims := &models.Claims{
		UserID:   "test-user-123",
		Email:    "test@example.com",
		Provider: "google",
		Name:     "Test User",
		Avatar:   "https://example.com/avatar.png",
	}

	token, err := service.GenerateJWT(claims)

	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.True(t, claims.ExpiresAt > time.Now().Unix())
	assert.True(t, claims.ExpiresAt <= time.Now().Add(24*time.Hour).Unix())
}

func TestGenerateJWT_ValidTokenParsing(t *testing.T) {
	cfg := config.Load()
	cfg.JWTSecret = "test-secret-key"
	service := NewService(cfg, nil)

	claims := &models.Claims{
		UserID:   "test-user-456",
		Email:    "user@test.com",
		Provider: "github",
		Name:     "GitHub User",
	}

	tokenString, err := service.GenerateJWT(claims)
	require.NoError(t, err)

	token, err := jwt.ParseWithClaims(tokenString, &models.Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.JWTSecret), nil
	})

	require.NoError(t, err)
	assert.True(t, token.Valid)

	parsedClaims := token.Claims.(*models.Claims)
	assert.Equal(t, "test-user-456", parsedClaims.UserID)
	assert.Equal(t, "user@test.com", parsedClaims.Email)
	assert.Equal(t, "github", parsedClaims.Provider)
}

func TestNewService(t *testing.T) {
	cfg := config.Load()
	service := NewService(cfg, nil)

	assert.NotNil(t, service)
	assert.NotNil(t, service.googleCfg)
	assert.NotNil(t, service.githubCfg)
	assert.Equal(t, cfg.GoogleClientID, service.googleCfg.ClientID)
	assert.Equal(t, cfg.GitHubClientID, service.githubCfg.ClientID)
	assert.Equal(t, cfg.GoogleClientSecret, service.googleCfg.ClientSecret)
	assert.Equal(t, cfg.GitHubClientSecret, service.githubCfg.ClientSecret)
}

func TestGoogleAuthURL(t *testing.T) {
	cfg := config.Load()
	service := NewService(cfg, nil)

	url := service.GoogleAuthURL()

	assert.NotEmpty(t, url)
	assert.Contains(t, url, "accounts.google.com")
	assert.Contains(t, url, "client_id=")
	assert.Contains(t, url, "redirect_uri=")
	assert.Contains(t, url, "scope=")
	assert.Contains(t, url, "response_type=code")
}

func TestGitHubAuthURL(t *testing.T) {
	cfg := config.Load()
	service := NewService(cfg, nil)

	url := service.GitHubAuthURL()

	assert.NotEmpty(t, url)
	assert.Contains(t, url, "github.com")
	assert.Contains(t, url, "client_id=")
	assert.Contains(t, url, "redirect_uri=")
	assert.Contains(t, url, "scope=")
	assert.Contains(t, url, "response_type=code")
}

func TestExchangeGoogleCode_InvalidCode(t *testing.T) {
	cfg := config.Load()
	service := NewService(cfg, nil)

	token, err := service.ExchangeGoogleCode("invalid-code")

	assert.Error(t, err)
	assert.Nil(t, token)
}

func TestExchangeGitHubCode_InvalidCode(t *testing.T) {
	cfg := config.Load()
	service := NewService(cfg, nil)

	token, err := service.ExchangeGitHubCode("invalid-code")

	assert.Error(t, err)
	assert.Nil(t, token)
}

func TestGetGoogleUserInfo_InvalidToken(t *testing.T) {
	cfg := config.Load()
	service := NewService(cfg, nil)

	token := &oauth2.Token{AccessToken: "invalid-token"}
	claims, err := service.GetGoogleUserInfo(token)

	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Empty(t, claims.UserID)
	assert.Empty(t, claims.Email)
}

func TestGetGitHubUserInfo_InvalidToken(t *testing.T) {
	cfg := config.Load()
	service := NewService(cfg, nil)

	token := &oauth2.Token{AccessToken: "invalid-token"}
	claims, err := service.GetGitHubUserInfo(token)

	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, "github:0", claims.UserID)
}

func TestFindOrCreateUser_WithoutRepo(t *testing.T) {
	cfg := config.Load()
	service := NewService(cfg, nil)

	_, err := service.FindOrCreateUser(context.Background(), "google", "123", "test@example.com", "Test", "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not configured")
}
