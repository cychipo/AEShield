package files

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"mime"
	"mime/multipart"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/aeshield/backend/internal/crypto"
	"github.com/aeshield/backend/internal/storage"
	"github.com/aeshield/backend/models"
	"github.com/gofiber/fiber/v2"
)

type JobHandler struct {
	jobService *JobService
	fileSvc    *Service
}

func NewJobHandler(jobService *JobService, fileSvc *Service) *JobHandler {
	return &JobHandler{jobService: jobService, fileSvc: fileSvc}
}

func (h *JobHandler) Create(c *fiber.Ctx) error {
	claims := getUserClaims(c)
	if claims == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	jobType := c.FormValue("type")
	if jobType == "" {
		jobType = string(models.JobTypeEncrypt)
	}
	password := c.FormValue("password")
	if err := ensurePasswordForJob(password); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if models.JobType(jobType) == models.JobTypeDecrypt {
		return h.createDecryptJob(c, claims.UserID, password)
	}
	if models.JobType(jobType) != models.JobTypeEncrypt {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "unsupported job type"})
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "file is required"})
	}

	job, err := h.jobService.CreateJob(c.Context(), claims.UserID, models.JobTypeEncrypt, fileHeader.Filename)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	if err := h.jobService.Start(c.Context(), job.ID.Hex()); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	jobCtx, cancel := context.WithTimeout(context.Background(), JobExecutionTimeout)
	go h.runEncryptJob(jobCtx, cancel, claims.UserID, job.ID.Hex(), fileHeader, password, c.FormValue("encryption_type", string(models.EncryptionAES256)), c.FormValue("access_mode", string(models.AccessModePrivate)))

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{"job_id": job.ID.Hex(), "status": job.Status, "progress": job.Progress})
}

func (h *JobHandler) createDecryptJob(c *fiber.Ctx, userID, password string) error {
	fileID := strings.TrimSpace(c.FormValue("file_id"))
	if fileID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "file_id is required"})
	}

	file, err := h.fileSvc.fileRepo.FindByID(c.Context(), fileID)
	if err != nil {
		if err == storage.ErrFileNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "file not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if !canAccessFile(file, userID) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "access denied"})
	}

	job, err := h.jobService.CreateJob(c.Context(), userID, models.JobTypeDecrypt, file.Filename)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if err := h.jobService.Start(c.Context(), job.ID.Hex()); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	jobCtx, cancel := context.WithTimeout(context.Background(), JobExecutionTimeout)
	go h.runDecryptJob(jobCtx, cancel, userID, job.ID.Hex(), file, password)
	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{"job_id": job.ID.Hex(), "status": models.JobStatusProcessing, "progress": 0})
}

func (h *JobHandler) runEncryptJob(ctx context.Context, cancel context.CancelFunc, userID, jobID string, fileHeader *multipart.FileHeader, password, encryptionType, accessMode string) {
	defer cancel()

	src, err := fileHeader.Open()
	if err != nil {
		_ = h.jobService.Fail(ctx, jobID, "open_file_failed", "cannot open file")
		return
	}
	defer src.Close()

	result, err := h.fileSvc.Upload(ctx, UploadInput{
		OwnerID:        userID,
		Filename:       fileHeader.Filename,
		Size:           fileHeader.Size,
		ContentType:    fileHeader.Header.Get("Content-Type"),
		EncryptionType: encryptionType,
		AccessMode:     accessMode,
		Password:       password,
		Body:           src,
		JobID:          jobID,
	})
	if err != nil {
		if errors.Is(err, ErrJobCancelled) {
			return
		}
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
			_ = h.jobService.Fail(context.Background(), jobID, "timeout", "Tác vụ mã hóa đã vượt quá thời gian cho phép.")
			return
		}
		_ = h.jobService.Fail(ctx, jobID, "upload_failed", err.Error())
		return
	}

	_ = h.jobService.Complete(ctx, jobID, &models.JobResult{FileID: result.ID.Hex(), Filename: result.Filename}, nil)
}

func (h *JobHandler) runDecryptJob(ctx context.Context, cancel context.CancelFunc, userID, jobID string, file *models.FileMetadata, password string) {
	defer cancel()

	reader, err := h.fileSvc.r2.GetFile(ctx, file.StoragePath)
	if err != nil {
		_ = h.jobService.Fail(ctx, jobID, "download_failed", err.Error())
		return
	}
	defer reader.Close()

	var output bytes.Buffer
	err = crypto.DecryptWithProgress(&output, reader, password, func(processedBytes int64) error {
		cancelled, err := h.jobService.IsCancelled(ctx, jobID)
		if err != nil {
			return err
		}
		if cancelled {
			return ErrJobCancelled
		}
		return h.jobService.UpdateProgress(ctx, jobID, progressFromBytes(processedBytes, file.Size))
	})
	if err != nil {
		if errors.Is(err, ErrJobCancelled) {
			return
		}
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
			_ = h.jobService.Fail(context.Background(), jobID, "timeout", "Tác vụ giải mã đã vượt quá thời gian cho phép.")
			return
		}
		code := "decrypt_failed"
		message := err.Error()
		if errors.Is(err, crypto.ErrDecryptChunk) {
			code = "wrong_password"
			message = "Sai mật khẩu hoặc tệp mã hóa không hợp lệ."
		}
		_ = h.jobService.Fail(ctx, jobID, code, message)
		return
	}

	previewType, mimeType := detectDecryptPreview(file.Filename)
	result := &models.JobResult{
		FileID:        file.ID.Hex(),
		Filename:      file.Filename,
		PreviewType:   previewType,
		MimeType:      mimeType,
		ContentBase64: base64.StdEncoding.EncodeToString(output.Bytes()),
	}
	_ = h.jobService.Complete(ctx, jobID, result, nil)
}

func detectDecryptPreview(filename string) (string, string) {
	ext := strings.ToLower(filepath.Ext(filename))
	mimeType := mime.TypeByExtension(ext)
	switch ext {
	case ".txt", ".md", ".json", ".csv", ".log", ".xml", ".js", ".jsx", ".ts", ".tsx", ".html", ".css", ".go", ".py", ".java", ".sql", ".yaml", ".yml":
		return "text", fallbackMime(mimeType, "text/plain")
	case ".png", ".jpg", ".jpeg", ".gif", ".webp", ".bmp", ".svg":
		return "image", fallbackMime(mimeType, "image/*")
	case ".pdf":
		return "pdf", fallbackMime(mimeType, "application/pdf")
	default:
		return "unsupported", fallbackMime(mimeType, "application/octet-stream")
	}
}

func fallbackMime(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func canAccessFile(file *models.FileMetadata, userID string) bool {
	switch models.AccessMode(file.AccessMode) {
	case models.AccessModePublic:
		return true
	case models.AccessModePrivate:
		return file.OwnerID == userID
	case models.AccessModeWhitelist:
		return file.OwnerID == userID || contains(file.Whitelist, userID)
	default:
		return false
	}
}

func (h *JobHandler) Get(c *fiber.Ctx) error {
	claims := getUserClaims(c)
	if claims == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	job, err := h.jobService.GetJob(c.Context(), c.Params("id"))
	if err != nil {
		if err == storage.ErrJobNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "job not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if job.UserID != claims.UserID {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "job not found"})
	}
	return c.JSON(job)
}

func (h *JobHandler) List(c *fiber.Ctx) error {
	claims := getUserClaims(c)
	if claims == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	status := c.Query("status")
	limit, _ := strconv.ParseInt(c.Query("limit", "20"), 10, 64)
	offset, _ := strconv.ParseInt(c.Query("offset", "0"), 10, 64)
	if limit <= 0 {
		limit = 20
	}
	jobs, total, err := h.jobService.ListJobs(c.Context(), claims.UserID, status, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"jobs": jobs, "total": total, "limit": limit, "offset": offset})
}

func (h *JobHandler) Cancel(c *fiber.Ctx) error {
	claims := getUserClaims(c)
	if claims == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	job, err := h.jobService.GetJob(c.Context(), c.Params("id"))
	if err != nil {
		if err == storage.ErrJobNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "job not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if job.UserID != claims.UserID {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "job not found"})
	}
	if job.Status == models.JobStatusCompleted || job.Status == models.JobStatusFailed || job.Status == models.JobStatusCancelled {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "job is not running"})
	}
	if err := h.jobService.Cancel(c.Context(), c.Params("id")); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"job_id": c.Params("id"), "status": models.JobStatusCancelled})
}

func cleanupJobsLoop(ctx context.Context, jobService *JobService) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_ = jobService.CleanupOldJobs(context.Background(), 24*time.Hour)
		}
	}
}

