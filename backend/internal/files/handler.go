package files

import (
	"log"

	"github.com/aeshield/backend/internal/storage"
	"github.com/aeshield/backend/models"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Upload godoc
//
//	@Summary		Upload file
//	@Description	Upload và lưu metadata file (stream mã hóa AES-GCM, không buffer toàn bộ vào RAM)
//	@Tags			files
//	@Accept			multipart/form-data
//	@Produce		json
//	@Security		BearerAuth
//	@Param			file			formData	file	true	"File cần upload"
//	@Param			password		formData	string	true	"Mật khẩu mã hóa file"
//	@Param			encryption_type	formData	string	false	"AES-128 | AES-192 | AES-256"
//	@Param			access_mode		formData	string	false	"public | private | whitelist"
//	@Success		201				{object}	models.FileMetadata
//	@Failure		400				{object}	map[string]string
//	@Failure		401				{object}	map[string]string
//	@Failure		500				{object}	map[string]string
//	@Router			/files/upload [post]
func (h *Handler) Upload(c *fiber.Ctx) error {
	claims := getUserClaims(c)
	if claims == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "file is required"})
	}

	password := c.FormValue("password")
	if password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "password is required"})
	}

	encryptionType := c.FormValue("encryption_type", string(models.EncryptionAES256))
	accessMode := c.FormValue("access_mode", string(models.AccessModePrivate))

	// Validate encryption_type
	switch models.EncryptionType(encryptionType) {
	case models.EncryptionAES128, models.EncryptionAES192, models.EncryptionAES256:
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid encryption_type"})
	}

	// Validate access_mode
	switch models.AccessMode(accessMode) {
	case models.AccessModePublic, models.AccessModePrivate, models.AccessModeWhitelist:
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid access_mode"})
	}

	src, err := fileHeader.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "cannot open file"})
	}
	defer src.Close()

	log.Printf("[upload.debug] handler receive owner=%s filename=%q size=%d contentType=%q encryption=%s access=%s", claims.UserID, fileHeader.Filename, fileHeader.Size, fileHeader.Header.Get("Content-Type"), encryptionType, accessMode)

	result, err := h.service.Upload(c.Context(), UploadInput{
		OwnerID:        claims.UserID,
		Filename:       fileHeader.Filename,
		Size:           fileHeader.Size,
		ContentType:    fileHeader.Header.Get("Content-Type"),
		EncryptionType: encryptionType,
		AccessMode:     accessMode,
		Password:       password,
		Body:           src,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(result)
}

// Download godoc
//
//	@Summary		Download file
//	@Description	Kiểm tra quyền và trả về presigned URL để download
//	@Tags			files
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"File ID"
//	@Success		200	{object}	map[string]string
//	@Failure		403	{object}	map[string]string
//	@Failure		404	{object}	map[string]string
//	@Router			/files/{id}/download [get]
func (h *Handler) Download(c *fiber.Ctx) error {
	claims := getUserClaims(c)
	requesterID := ""
	if claims != nil {
		requesterID = claims.UserID
	}

	fileID := c.Params("id")
	url, err := h.service.GetDownloadURL(c.Context(), fileID, requesterID)
	if err != nil {
		if err == storage.ErrFileNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "file not found"})
		}
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"url": url})
}

// Delete godoc
//
//	@Summary		Delete file
//	@Description	Xóa file trên R2 và metadata trong MongoDB
//	@Tags			files
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"File ID"
//	@Success		200	{object}	map[string]string
//	@Failure		403	{object}	map[string]string
//	@Failure		404	{object}	map[string]string
//	@Router			/files/{id} [delete]
func (h *Handler) Delete(c *fiber.Ctx) error {
	claims := getUserClaims(c)
	if claims == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	fileID := c.Params("id")
	if err := h.service.Delete(c.Context(), fileID, claims.UserID); err != nil {
		if err == storage.ErrFileNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "file not found"})
		}
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "deleted"})
}

// ListFiles godoc
//
//	@Summary		List files
//	@Description	Lấy danh sách file của user hiện tại
//	@Tags			files
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{array}		models.FileMetadata
//	@Router			/files [get]
func (h *Handler) ListFiles(c *fiber.Ctx) error {
	claims := getUserClaims(c)
	if claims == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	fileList, err := h.service.ListFiles(c.Context(), claims.UserID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fileList)
}

// Share godoc
//
//	@Summary		Share file
//	@Description	Cập nhật access mode và/hoặc whitelist của file (chỉ owner)
//	@Tags			files
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		shareRequest	true	"Share settings"
//	@Success		200		{object}	models.FileMetadata
//	@Failure		400		{object}	map[string]string
//	@Failure		403		{object}	map[string]string
//	@Failure		404		{object}	map[string]string
//	@Router			/files/share [patch]
func (h *Handler) Share(c *fiber.Ctx) error {
	claims := getUserClaims(c)
	if claims == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	var req shareRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if req.FileID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "file_id is required"})
	}

	file, err := h.service.Share(c.Context(), ShareInput{
		FileID:     req.FileID,
		OwnerID:    claims.UserID,
		AccessMode: req.AccessMode,
		Whitelist:  req.Whitelist,
	})
	if err != nil {
		if err == storage.ErrFileNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "file not found"})
		}
		if err.Error() == "access denied" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(file)
}

type shareRequest struct {
	FileID     string   `json:"file_id"`
	AccessMode string   `json:"access_mode"`
	Whitelist  []string `json:"whitelist"`
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
