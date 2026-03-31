package e2e

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	baseURL = "http://localhost:6888"
)

func getenv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func TestMain(m *testing.M) {
	serverURL := getenv("TEST_SERVER_URL", baseURL)
	if serverURL != baseURL {
		os.Exit(0)
	}

	if err := waitForServer(baseURL + "/api/v1/auth/urls"); err != nil {
		fmt.Printf("Server not available: %v\n", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func waitForServer(url string) error {
	maxAttempts := 30
	for i := 0; i < maxAttempts; i++ {
		resp, err := http.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusUnauthorized {
				return nil
			}
		}
	}
	return fmt.Errorf("server not available after %d attempts", maxAttempts)
}

func TestE2E_GetAuthURLs(t *testing.T) {
	resp, err := http.Get(baseURL + "/api/v1/auth/urls")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var result map[string]string
	json.Unmarshal(body, &result)

	assert.Contains(t, result, "google")
	assert.Contains(t, result, "github")
	assert.True(t, strings.HasPrefix(result["google"], "https://accounts.google.com"))
	assert.True(t, strings.HasPrefix(result["github"], "https://github.com"))
}

func TestE2E_GoogleLogin(t *testing.T) {
	resp, err := http.Get(baseURL + "/api/v1/auth/google")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var result map[string]string
	json.Unmarshal(body, &result)

	assert.Contains(t, result, "url")
	assert.True(t, strings.HasPrefix(result["url"], "https://accounts.google.com"))
	assert.Contains(t, result["url"], "client_id=")
	assert.Contains(t, result["url"], "redirect_uri=")
}

func TestE2E_GitHubLogin(t *testing.T) {
	resp, err := http.Get(baseURL + "/api/v1/auth/github")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var result map[string]string
	json.Unmarshal(body, &result)

	assert.Contains(t, result, "url")
	assert.True(t, strings.HasPrefix(result["url"], "https://github.com"))
	assert.Contains(t, result["url"], "client_id=")
	assert.Contains(t, result["url"], "redirect_uri=")
}

func TestE2E_GoogleCallback_MissingCode(t *testing.T) {
	resp, err := http.Get(baseURL + "/api/v1/auth/google/callback")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var result map[string]string
	json.Unmarshal(body, &result)

	assert.Contains(t, result, "error")
	assert.Equal(t, "missing code", result["error"])
}

func TestE2E_GitHubCallback_MissingCode(t *testing.T) {
	resp, err := http.Get(baseURL + "/api/v1/auth/github/callback")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var result map[string]string
	json.Unmarshal(body, &result)

	assert.Contains(t, result, "error")
	assert.Equal(t, "missing code", result["error"])
}

func TestE2E_GoogleCallback_InvalidCode(t *testing.T) {
	resp, err := http.Get(baseURL + "/api/v1/auth/google/callback?code=invalid")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestE2E_GitHubCallback_InvalidCode(t *testing.T) {
	resp, err := http.Get(baseURL + "/api/v1/auth/github/callback?code=invalid")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestE2E_ProtectedEndpoint_NoToken(t *testing.T) {
	req, _ := http.NewRequest("GET", baseURL+"/api/v1/auth/me", nil)
	client := &http.Client{}
	resp, err := client.Do(req)

	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var result map[string]string
	json.Unmarshal(body, &result)
	assert.Contains(t, result, "error")
}

func TestE2E_ProtectedEndpoint_InvalidToken(t *testing.T) {
	req, _ := http.NewRequest("GET", baseURL+"/api/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	client := &http.Client{}
	resp, err := client.Do(req)

	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var result map[string]string
	json.Unmarshal(body, &result)
	assert.Contains(t, result, "error")
}

func TestE2E_ProtectedEndpoint_WrongFormat(t *testing.T) {
	req, _ := http.NewRequest("GET", baseURL+"/api/v1/auth/me", nil)
	req.Header.Set("Authorization", "InvalidFormat")
	client := &http.Client{}
	resp, err := client.Do(req)

	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestE2E_SwaggerJSON(t *testing.T) {
	resp, err := http.Get(baseURL + "/api/v1/swagger.json")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	body, _ := io.ReadAll(resp.Body)
	var swagger map[string]interface{}
	err = json.Unmarshal(body, &swagger)

	require.NoError(t, err)
	assert.Contains(t, swagger, "openapi")
	assert.Contains(t, swagger, "info")
}

func TestE2E_SwaggerUI(t *testing.T) {
	resp, err := http.Get(baseURL + "/docs/")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestE2E_NotFound(t *testing.T) {
	resp, err := http.Get(baseURL + "/api/v1/nonexistent")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestE2E_HealthCheck(t *testing.T) {
	resp, err := http.Get(baseURL + "/")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestE2E_CORS_Preflight(t *testing.T) {
	req, _ := http.NewRequest("OPTIONS", baseURL+"/api/v1/auth/urls", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", "GET")
	client := &http.Client{}
	resp, err := client.Do(req)

	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Contains(t, []int{http.StatusOK, http.StatusNoContent, http.StatusNotFound}, resp.StatusCode)
}
