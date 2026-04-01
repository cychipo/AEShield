package auth

import (
	"errors"
	"fmt"

	"github.com/aeshield/backend/internal/database"
	"github.com/gofiber/fiber/v2"
)

// GetAuthURLs
//
//	@Summary		Get OAuth URLs
//	@Description	Get Google and GitHub OAuth authorization URLs
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	map[string]string
//	@Router			/auth/urls [get]
func (h *Handler) GetAuthURLs(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"google": h.service.GoogleAuthURL(),
		"github": h.service.GitHubAuthURL(),
	})
}

// GoogleLogin
//
//	@Summary		Google OAuth Login
//	@Description	Redirect to Google OAuth authorization page
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	map[string]string
//	@Router			/auth/google [get]
func (h *Handler) GoogleLogin(c *fiber.Ctx) error {
	url := h.service.GoogleAuthURL()
	return c.JSON(fiber.Map{"url": url})
}

// GitHubLogin
//
//	@Summary		GitHub OAuth Login
//	@Description	Redirect to GitHub OAuth authorization page
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	map[string]string
//	@Router			/auth/github [get]
func (h *Handler) GitHubLogin(c *fiber.Ctx) error {
	url := h.service.GitHubAuthURL()
	return c.JSON(fiber.Map{"url": url})
}

// GoogleCallback
//
//	@Summary		Google OAuth Callback
//	@Description	Exchange Google authorization code for JWT token and create/update user
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			code	query		string	true	"Authorization code"
//	@Success		200		{object}	map[string]interface{}
//	@Failure		400		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/auth/google/callback [get]
func (h *Handler) GoogleCallback(c *fiber.Ctx) error {
	code := c.Query("code")
	if code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "missing code"})
	}

	oauthToken, err := h.service.ExchangeGoogleCode(code)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	claims, err := h.service.GetGoogleUserInfo(oauthToken)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	user, err := h.service.FindOrCreateUser(c.Context(), "google", claims.UserID, claims.Email, claims.Name, claims.Avatar)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	claims.UserID = user.ID.Hex()

	jwtToken, err := h.service.GenerateJWT(claims)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Redirect(fmt.Sprintf("%s/auth/google/callback?token=%s", h.service.cfg.FrontendURL, jwtToken), fiber.StatusTemporaryRedirect)
}

// GitHubCallback
//
//	@Summary		GitHub OAuth Callback
//	@Description	Exchange GitHub authorization code for JWT token and create/update user
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			code	query		string	true	"Authorization code"
//	@Success		200		{object}	map[string]interface{}
//	@Failure		400		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/auth/github/callback [get]
func (h *Handler) GitHubCallback(c *fiber.Ctx) error {
	code := c.Query("code")
	if code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "missing code"})
	}

	oauthToken, err := h.service.ExchangeGitHubCode(code)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	claims, err := h.service.GetGitHubUserInfo(oauthToken)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	user, err := h.service.FindOrCreateUser(c.Context(), "github", claims.UserID, claims.Email, claims.Name, claims.Avatar)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	claims.UserID = user.ID.Hex()

	jwtToken, err := h.service.GenerateJWT(claims)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Redirect(fmt.Sprintf("%s/auth/github/callback?token=%s", h.service.cfg.FrontendURL, jwtToken), fiber.StatusTemporaryRedirect)
}

// GetCurrentUser
//
//	@Summary		Get Current User
//	@Description	Get authenticated user information
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	models.Claims
//	@Router			/auth/me [get]
func (h *Handler) Me(c *fiber.Ctx) error {
	user := c.Locals("user")
	return c.JSON(user)
}

func (h *Handler) LookupUserByEmail(c *fiber.Ctx) error {
	result, err := h.service.LookupUserByEmail(c.Context(), c.Query("email"))
	if err != nil {
		if err.Error() == "email is required" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		if errors.Is(err, database.ErrUserNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(result)
}

func (h *Handler) ResolveUsersByID(c *fiber.Ctx) error {
	var req UserResolveRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	result, err := h.service.ResolveUsersByID(c.Context(), req.IDs)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(result)
}

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}
