package notifications

import (
	"context"
	"strconv"

	"github.com/aeshield/backend/models"
	"github.com/gofiber/fiber/v2"
)

type NotificationService interface {
	List(ctx context.Context, recipientUserID string, limit int64, cursor string) (*ListResponse, error)
	MarkAllRead(ctx context.Context, recipientUserID string) (*MarkAllReadResponse, error)
}

type Handler struct {
	service NotificationService
}

func NewHandler(service NotificationService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) List(c *fiber.Ctx) error {
	claims := getUserClaims(c)
	if claims == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	limit := int64(5)
	if rawLimit := c.Query("limit"); rawLimit != "" {
		parsedLimit, err := strconv.ParseInt(rawLimit, 10, 64)
		if err != nil || parsedLimit <= 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid limit"})
		}
		if parsedLimit > 50 {
			parsedLimit = 50
		}
		limit = parsedLimit
	}

	response, err := h.service.List(c.Context(), claims.UserID, limit, c.Query("cursor"))
	if err != nil {
		if err.Error() == "invalid cursor" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(response)
}

func (h *Handler) MarkAllRead(c *fiber.Ctx) error {
	claims := getUserClaims(c)
	if claims == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	response, err := h.service.MarkAllRead(c.Context(), claims.UserID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(response)
}

func getUserClaims(c *fiber.Ctx) *models.Claims {
	user := c.Locals("user")
	if user == nil {
		return nil
	}
	claims, ok := user.(*models.Claims)
	if !ok {
		return nil
	}
	return claims
}
