# Testing Guide

## Overview

This document describes the testing strategy and guidelines for the AEShield backend.

## Test Types

### 1. Unit Tests (`*_test.go`)

- **Location**: Same package as the code being tested
- **Purpose**: Test individual functions and methods in isolation
- **Dependencies**: Mocked or using test configurations
- **Run**: `go test ./internal/auth/... -v`

### 2. End-to-End Tests (`e2e/*_test.go`)

- **Location**: `e2e/` directory
- **Purpose**: Test full HTTP request/response flow against running server
- **Dependencies**: Requires running server
- **Run**: `go test ./e2e/... -v`

## Running Tests

### Unit Tests

```bash
# Run all unit tests
cd backend
go test ./... -v

# Run specific package
go test ./internal/auth/... -v

# Run with coverage
go test ./... -cover

# Run specific test
go test ./internal/auth/... -run TestGetAuthURLs -v
```

### End-to-End Tests

```bash
# Start the server first
cd backend
./server &

# Run e2e tests (in another terminal)
go test ./e2e/... -v

# Or set custom server URL
TEST_SERVER_URL=http://localhost:6888 go test ./e2e/... -v

# Default port is 6888
```

### All Tests

```bash
# Run both unit and e2e tests
go test ./... -v

# Skip e2e tests (requires server)
go test ./... -v -skip "E2E"
```

## Using Mock Repository

For testing services that depend on database repositories, use mock implementations:

```go
// Mock User Repository
type MockUserRepository struct {
    users map[string]*models.User
}

func NewMockUserRepository() *MockUserRepository {
    return &MockUserRepository{
        users: make(map[string]*models.User),
    }
}

func (m *MockUserRepository) FindByProvider(ctx context.Context, provider, providerID string) (*models.User, error) {
    key := provider + ":" + providerID
    user, ok := m.users[key]
    if !ok {
        return nil, ErrUserNotFound
    }
    return user, nil
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
    // implementation
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
    // implementation
}

func (m *MockUserRepository) Update(ctx context.Context, user *models.User) error {
    // implementation
}

// In test
mockRepo := NewMockUserRepository()
service := NewService(cfg, mockRepo)
```

## Adding Tests for New Modules

### 1. Create Unit Test

Create `*_test.go` in the same package:

```go
package auth

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestYourFunction(t *testing.T) {
    // Arrange
    // ...

    // Act
    result, err := YourFunction()

    // Assert
    require.NoError(t, err)
    assert.Equal(t, expected, result)
}
```

### 2. Create E2E Test

Create `e2e/<module>_e2e_test.go`:

```go
package e2e

import (
    "testing"
    "net/http"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

const baseURL = "http://localhost:6888"

func TestE2E_YourEndpoint(t *testing.T) {
    resp, err := http.Get(baseURL + "/api/v1/your/endpoint")
    require.NoError(t, err)
    defer resp.Body.Close()

    assert.Equal(t, http.StatusOK, resp.StatusCode)
}
```

### 3. Test Naming Conventions

| Pattern | Description |
|---------|-------------|
| `Test<Module><Method>` | Handler/Service tests |
| `Test<Module>_<Scenario>` | E2E tests |
| `Test<Module>_<ErrorCase>_Error` | Error case tests |
| `Benchmark<Module><Method>` | Benchmark tests |

### 4. Test Structure

```go
func TestYourFeature(t *testing.T) {
    // Arrange - Setup test data and dependencies
    // ...
    
    // Act - Execute the function being tested
    result, err := YourFunction()
    
    // Assert - Verify the results
    require.NoError(t, err)
    assert.Equal(t, expected, result)
    assert.NotNil(t, result.ID)
}
```

## Auth Service Tests Example

### Unit Tests with Mock

```go
func TestFindOrCreateUser_CreateNew(t *testing.T) {
    cfg := config.Load()
    mockRepo := NewMockUserRepository()
    service := NewService(cfg, mockRepo)

    user, err := service.FindOrCreateUser(
        context.Background(),
        "google",
        "google-123",
        "test@example.com",
        "Test User",
        "https://example.com/avatar.jpg",
    )

    require.NoError(t, err)
    assert.NotNil(t, user)
    assert.Equal(t, "google", user.Provider)
    assert.Equal(t, "test@example.com", user.Email)
}

func TestFindOrCreateUser_UpdateExisting(t *testing.T) {
    cfg := config.Load()
    mockRepo := NewMockUserRepository()

    // Create existing user
    existingUser := &models.User{
        Provider:   "google",
        ProviderID: "google-123",
        Email:      "old@example.com",
        Name:       "Old Name",
    }
    mockRepo.Create(context.Background(), existingUser)

    service := NewService(cfg, mockRepo)

    // Update with new info
    user, err := service.FindOrCreateUser(
        context.Background(),
        "google",
        "google-123",
        "new@example.com",
        "New Name",
        "https://example.com/new.jpg",
    )

    require.NoError(t, err)
    assert.Equal(t, "new@example.com", user.Email)
    assert.Equal(t, "New Name", user.Name)
}
```

### JWT Middleware Tests

```go
func TestJWTMiddlewareValidToken(t *testing.T) {
    app := fiber.New()
    secret := "test-secret"

    app.Get("/protected", JWTMiddleware(secret), func(c *fiber.Ctx) error {
        user := c.Locals("user").(*models.Claims)
        return c.JSON(user)
    })

    claims := &models.Claims{
        UserID:    "test-user-123",
        Email:     "test@example.com",
        Provider:  "google",
        ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, _ := token.SignedString([]byte(secret))

    req := httptest.NewRequest("GET", "/protected", nil)
    req.Header.Set("Authorization", "Bearer "+tokenString)
    resp, err := app.Test(req)

    require.NoError(t, err)
    assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}
```

## Test Dependencies

We use [testify](https://github.com/stretchr/testify) for assertions:

```bash
go get github.com/stretchr/testify
```

### Common Assertions

```go
import (
    "github.com/stretchr/testify/assert"   // assertions
    "github.com/stretchr/testify/require"  // required assertions
)

// Basic
assert.Equal(t, expected, actual)
assert.NotEqual(t, expected, actual)
assert.Nil(t, value)
assert.NotNil(t, value)
assert.True(t, condition)
assert.False(t, condition)

// Error handling
require.NoError(t, err)
assert.Error(t, err)

// Slices/Maps
assert.Contains(t, slice, element)
assert.Contains(t, map, key)

// Strings
assert.Contains(t, string, substring)
assert.HasPrefix(t, string, prefix)
assert.Matches(t, string, regex)
```

## CI/CD Integration

Add to your CI pipeline:

```yaml
# .github/workflows/test.yml
- name: Run tests
  run: |
    cd backend
    go test ./... -v -race -cover

- name: Run e2e tests
  run: |
    cd backend
    ./server &
    sleep 5
    go test ./e2e/... -v
```

## Notes

- E2E tests use port 6888 by default (configurable via TEST_SERVER_URL)
- E2E tests are skipped if server is not available
- Use `TestMain` for setup/teardown logic
- Keep tests independent - no shared state between tests
- Use descriptive test names that explain the scenario
- Use mocks for database dependencies in unit tests
- Skip tests that require external APIs (GitHub, Google) in CI
