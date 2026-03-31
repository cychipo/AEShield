// Package handler cung cấp các endpoint HTTP cho hệ thống quản lý quyền truy cập
package handler

import (
	acmodels "github.com/aeshield/backend/internal/accesscontrol/models"
	"github.com/aeshield/backend/internal/accesscontrol/service"
	bmodels "github.com/aeshield/backend/models"
	"github.com/gofiber/fiber/v2"
)

// Handler định nghĩa các phương thức xử lý HTTP cho quản lý quyền truy cập
type Handler struct {
	service *service.AccessControlService
}

// NewHandler tạo mới một instance của Handler
func NewHandler(service *service.AccessControlService) *Handler {
	return &Handler{
		service: service,
	}
}

// CreateRule godoc
//
//	@Summary		Tạo mới quy tắc truy cập
//	@Description	Tạo quy tắc truy cập mới cho tài nguyên
//	@Tags			access-control
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		CreateRuleRequest	true	"Thông tin quy tắc truy cập"
//	@Success		201		{object}	acmodels.AccessRule
//	@Failure		400		{object}	map[string]string
//	@Failure		401		{object}	map[string]string
//	@Failure		403		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/access/rules [post]
func (h *Handler) CreateRule(c *fiber.Ctx) error {
	claims := getUserClaims(c)
	if claims == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	var req CreateRuleRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	// Validate access mode
	switch acmodels.AccessMode(req.AccessMode) {
	case acmodels.AccessModePublic, acmodels.AccessModePrivate, acmodels.AccessModeWhitelist:
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid access_mode"})
	}

	// Check if user is allowed to create rule for this resource
	// In a real implementation, you might want to add additional checks here

	rule, err := h.service.CreateRule(
		c.Context(),
		req.ResourceID,
		req.ResourceType,
		claims.UserID,
		acmodels.AccessMode(req.AccessMode),
		req.Whitelist,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(rule)
}

// GetRule godoc
//
//	@Summary		Lấy quy tắc truy cập
//	@Description	Lấy thông tin quy tắc truy cập theo ID tài nguyên
//	@Tags			access-control
//	@Produce		json
//	@Param			resource_id	path		string	true	"ID của tài nguyên"
//	@Success		200			{object}	acmodels.AccessRule
//	@Failure		404			{object}	map[string]string
//	@Failure		500			{object}	map[string]string
//	@Router			/access/rules/{resource_id} [get]
func (h *Handler) GetRule(c *fiber.Ctx) error {
	resourceID := c.Params("resource_id")

	rule, err := h.service.GetRuleByResource(c.Context(), resourceID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "access rule not found"})
	}

	// Check if user has permission to view this rule
	claims := getUserClaims(c)
	if claims != nil {
		isOwner, _ := h.service.IsOwner(c.Context(), resourceID, claims.UserID)
		if !isOwner {
			// For non-owners, only return basic info (hide whitelist for security)
			rule.Whitelist = []string{}
		}
	}

	return c.JSON(rule)
}

// UpdateRule godoc
//
//	@Summary		Cập nhật quy tắc truy cập
//	@Description	Cập nhật quy tắc truy cập cho tài nguyên (chỉ chủ sở hữu)
//	@Tags			access-control
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			resource_id	path		string				true	"ID của tài nguyên"
//	@Param			body		body		UpdateRuleRequest	true	"Thông tin cập nhật quy tắc"
//	@Success		200			{object}	acmodels.AccessRule
//	@Failure		400			{object}	map[string]string
//	@Failure		401			{object}	map[string]string
//	@Failure		403			{object}	map[string]string
//	@Failure		404			{object}	map[string]string
//	@Failure		500			{object}	map[string]string
//	@Router			/access/rules/{resource_id} [patch]
func (h *Handler) UpdateRule(c *fiber.Ctx) error {
	claims := getUserClaims(c)
	if claims == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	resourceID := c.Params("resource_id")

	// Check if user is owner
	isOwner, err := h.service.IsOwner(c.Context(), resourceID, claims.UserID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to check ownership"})
	}
	if !isOwner {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "access denied"})
	}

	var req UpdateRuleRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	// Validate access mode
	switch acmodels.AccessMode(req.AccessMode) {
	case acmodels.AccessModePublic, acmodels.AccessModePrivate, acmodels.AccessModeWhitelist:
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid access_mode"})
	}

	rule, err := h.service.UpdateRule(
		c.Context(),
		resourceID,
		acmodels.AccessMode(req.AccessMode),
		req.Whitelist,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(rule)
}

// DeleteRule godoc
//
//	@Summary		Xóa quy tắc truy cập
//	@Description	Xóa quy tắc truy cập cho tài nguyên (chỉ chủ sở hữu)
//	@Tags			access-control
//	@Produce		json
//	@Security		BearerAuth
//	@Param			resource_id	path		string	true	"ID của tài nguyên"
//	@Success		200			{object}	map[string]string
//	@Failure		401			{object}	map[string]string
//	@Failure		403			{object}	map[string]string
//	@Failure		404			{object}	map[string]string
//	@Failure		500			{object}	map[string]string
//	@Router			/access/rules/{resource_id} [delete]
func (h *Handler) DeleteRule(c *fiber.Ctx) error {
	claims := getUserClaims(c)
	if claims == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	resourceID := c.Params("resource_id")

	// Check if user is owner
	isOwner, err := h.service.IsOwner(c.Context(), resourceID, claims.UserID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to check ownership"})
	}
	if !isOwner {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "access denied"})
	}

	err = h.service.DeleteRule(c.Context(), resourceID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "access rule deleted"})
}

// CheckAccess godoc
//
//	@Summary		Kiểm tra quyền truy cập
//	@Description	Kiểm tra xem người dùng có quyền truy cập tài nguyên hay không
//	@Tags			access-control
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		CheckAccessRequest	true	"Thông tin kiểm tra quyền truy cập"
//	@Success		200		{object}	acmodels.CheckAccessResult
//	@Failure		400		{object}	map[string]string
//	@Failure		401		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/access/check [post]
func (h *Handler) CheckAccess(c *fiber.Ctx) error {
	claims := getUserClaims(c)

	var req CheckAccessRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	request := &acmodels.CheckAccessRequest{
		ResourceID:    req.ResourceID,
		RequesterID:   "",
		RequesterEmail: "",
	}

	if claims != nil {
		request.RequesterID = claims.UserID
		request.RequesterEmail = claims.Email
	}

	result := h.service.CheckAccess(c.Context(), request)
	return c.JSON(result)
}

// AddToWhitelist godoc
//
//	@Summary		Thêm vào danh sách trắng
//	@Description	Thêm người dùng vào danh sách trắng của tài nguyên (chỉ chủ sở hữu)
//	@Tags			access-control
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		WhitelistRequest	true	"Thông tin thêm vào danh sách trắng"
//	@Success		200		{object}	acmodels.AccessRule
//	@Failure		400		{object}	map[string]string
//	@Failure		401		{object}	map[string]string
//	@Failure		403		{object}	map[string]string
//	@Failure		404		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/access/whitelist [post]
func (h *Handler) AddToWhitelist(c *fiber.Ctx) error {
	claims := getUserClaims(c)
	if claims == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	var req WhitelistRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	// Check if user is owner
	isOwner, err := h.service.IsOwner(c.Context(), req.ResourceID, claims.UserID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to check ownership"})
	}
	if !isOwner {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "access denied"})
	}

	err = h.service.AddToWhitelist(c.Context(), req.ResourceID, req.UserIdentifier)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Return updated rule
	rule, err := h.service.GetRuleByResource(c.Context(), req.ResourceID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(rule)
}

// RemoveFromWhitelist godoc
//
//	@Summary		Xóa khỏi danh sách trắng
//	@Description	Xóa người dùng khỏi danh sách trắng của tài nguyên (chỉ chủ sở hữu)
//	@Tags			access-control
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		WhitelistRequest	true	"Thông tin xóa khỏi danh sách trắng"
//	@Success		200		{object}	acmodels.AccessRule
//	@Failure		400		{object}	map[string]string
//	@Failure		401		{object}	map[string]string
//	@Failure		403		{object}	map[string]string
//	@Failure		404		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/access/whitelist [delete]
func (h *Handler) RemoveFromWhitelist(c *fiber.Ctx) error {
	claims := getUserClaims(c)
	if claims == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	var req WhitelistRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	// Check if user is owner
	isOwner, err := h.service.IsOwner(c.Context(), req.ResourceID, claims.UserID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to check ownership"})
	}
	if !isOwner {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "access denied"})
	}

	err = h.service.RemoveFromWhitelist(c.Context(), req.ResourceID, req.UserIdentifier)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Return updated rule
	rule, err := h.service.GetRuleByResource(c.Context(), req.ResourceID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(rule)
}

// Request structs
type CreateRuleRequest struct {
	ResourceID   string   `json:"resource_id"`
	ResourceType string   `json:"resource_type"`
	AccessMode   string   `json:"access_mode"`
	Whitelist    []string `json:"whitelist"`
}

type UpdateRuleRequest struct {
	AccessMode string   `json:"access_mode"`
	Whitelist  []string `json:"whitelist"`
}

type CheckAccessRequest struct {
	ResourceID string `json:"resource_id"`
}

type WhitelistRequest struct {
	ResourceID      string `json:"resource_id"`
	UserIdentifier  string `json:"user_identifier"`
}

// Helper function to get user claims
func getUserClaims(c *fiber.Ctx) *bmodels.Claims {
	user := c.Locals("user")
	if user == nil {
		return nil
	}
	claims, ok := user.(*bmodels.Claims)
	if !ok {
		return nil
	}
	return claims
}