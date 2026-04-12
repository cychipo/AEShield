package files

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/aeshield/backend/internal/crypto"
	"github.com/aeshield/backend/internal/storage"
	"github.com/aeshield/backend/models"
	"github.com/google/uuid"
)

// R2Storage là interface để có thể mock trong tests
type R2Storage interface {
	UploadFile(ctx context.Context, key string, body io.Reader, contentType string, size int64) error
	DeleteFile(ctx context.Context, key string) error
	GetFile(ctx context.Context, key string) (io.ReadCloser, error)
	GeneratePresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error)
}

// FileRepo là interface để có thể mock trong tests
type FileRepo interface {
	Create(ctx context.Context, file *models.FileMetadata) error
	FindByID(ctx context.Context, id string) (*models.FileMetadata, error)
	FindByOwner(ctx context.Context, ownerID string) ([]*models.FileMetadata, error)
	FindSharedWithUser(ctx context.Context, userID string) ([]*models.FileMetadata, error)
	Update(ctx context.Context, file *models.FileMetadata) error
	Delete(ctx context.Context, id string) error
}

type UserStorageRepo interface {
	GetByUserID(ctx context.Context, userID string) (*models.UserStorage, error)
	AdjustUsage(ctx context.Context, userID string, usedBytesDelta, fileCountDelta int64) error
	SetUsageIfEmpty(ctx context.Context, userID string, usedBytes, fileCount int64) error
}

var (
	ErrFileTooLarge = errors.New("file size exceeds 1GB limit")
	ErrStorageQuota = errors.New("storage quota exceeded")
)

type StorageUsageResponse struct {
	UsedBytes      int64   `json:"used_bytes"`
	QuotaBytes     int64   `json:"quota_bytes"`
	UsedGB         float64 `json:"used_gb"`
	QuotaGB        float64 `json:"quota_gb"`
	PercentUsed    float64 `json:"percent_used"`
	FileCount      int64   `json:"file_count"`
	AvailableBytes int64   `json:"available_bytes"`
}

type FileListResponse struct {
	OwnedFiles   []*models.FileMetadata `json:"owned_files"`
	SharedWithMe []*models.FileMetadata `json:"shared_with_me"`
}

type NotificationRepo interface {
	CreateMany(ctx context.Context, notifications []*models.Notification) error
}

type Service struct {
	r2               R2Storage
	fileRepo         FileRepo
	userStorageRepo  UserStorageRepo
	notificationRepo NotificationRepo
	jobService       *JobService
}

func NewService(r2 R2Storage, fileRepo FileRepo, userStorageRepo UserStorageRepo, notificationRepo NotificationRepo, jobService ...*JobService) *Service {
	var svc *JobService
	if len(jobService) > 0 {
		svc = jobService[0]
	}
	return &Service{
		r2:               r2,
		fileRepo:         fileRepo,
		userStorageRepo:  userStorageRepo,
		notificationRepo: notificationRepo,
		jobService:       svc,
	}
}

type UploadInput struct {
	OwnerID        string
	Filename       string
	Size           int64
	ContentType    string
	EncryptionType string
	AccessMode     string
	Password       string // Mật khẩu để mã hóa file (bắt buộc)
	Body           io.Reader
	JobID          string
}

// Upload stream mã hóa file và đẩy lên R2, lưu metadata vào MongoDB.
func (s *Service) Upload(ctx context.Context, input UploadInput) (*models.FileMetadata, error) {
	if input.Size > models.MaxFileSizeBytes {
		return nil, ErrFileTooLarge
	}

	usage, err := s.userStorageRepo.GetByUserID(ctx, input.OwnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to load storage usage: %w", err)
	}
	if usage.UsedBytes+input.Size > usage.QuotaBytes {
		return nil, ErrStorageQuota
	}

	// Sinh unique key cho R2
	ext := filepath.Ext(input.Filename)
	storageKey := fmt.Sprintf("%s/%s%s", input.OwnerID, uuid.New().String(), ext)

	// Map encryption type → KeyBits
	bits, err := encryptionTypeToBits(input.EncryptionType)
	if err != nil {
		return nil, err
	}

	// Validate password
	if input.Password == "" {
		return nil, fmt.Errorf("password is required for encryption")
	}

	// Pipe: mã hóa streaming → R2 upload
	pr, pw := io.Pipe()
	log.Printf("[upload.debug] service start owner=%s filename=%q plainSize=%d", input.OwnerID, input.Filename, input.Size)

	// Goroutine ghi ciphertext vào pipe writer
	var encryptErr error
	go func() {
		defer func() {
			if encryptErr != nil {
				pw.CloseWithError(encryptErr)
			} else {
				pw.Close()
			}
		}()
		_, encryptErr = crypto.EncryptWithProgress(pw, input.Body, input.Password, bits, func(processedBytes int64) error {
			if s.jobService != nil && input.JobID != "" {
				cancelled, err := s.jobService.IsCancelled(ctx, input.JobID)
				if err != nil {
					return err
				}
				if cancelled {
					return ErrJobCancelled
				}
				return s.jobService.UpdateProgress(ctx, input.JobID, progressFromBytes(processedBytes, input.Size))
			}
			return nil
		})
	}()

	// Upload lên R2 cần Content-Length chính xác và body có thể đọc ổn định.
	// Để tránh lỗi chữ ký với stream không seekable, buffer ciphertext trước khi PutObject.
	predictedEncryptedSize := encryptedSizeFromPlaintextSize(input.Size)
	encryptedData, err := io.ReadAll(pr)
	if err != nil {
		_ = pr.CloseWithError(err)
		return nil, fmt.Errorf("failed to prepare encrypted payload: %w", err)
	}
	encryptedSize := int64(len(encryptedData))
	log.Printf("[upload.debug] buffered mode plainSize=%d predictedEncryptedSize=%d actualEncryptedSize=%d filename=%q", input.Size, predictedEncryptedSize, encryptedSize, input.Filename)

	uploadBody := bytes.NewReader(encryptedData)
	if err := s.r2.UploadFile(ctx, storageKey, uploadBody, "application/octet-stream", encryptedSize); err != nil {
		log.Printf("[upload.debug] r2 upload error owner=%s filename=%q encryptedSize=%d err=%v", input.OwnerID, input.Filename, encryptedSize, err)
		pr.CloseWithError(err)
		return nil, fmt.Errorf("upload failed: %w", err)
	}
	log.Printf("[upload.debug] r2 upload success owner=%s filename=%q encryptedSize=%d storageKey=%q", input.OwnerID, input.Filename, encryptedSize, storageKey)
	if s.jobService != nil && input.JobID != "" {
		_ = s.jobService.UpdateProgress(ctx, input.JobID, 100)
	}

	// Tạo metadata
	file := &models.FileMetadata{
		OwnerID:        input.OwnerID,
		Filename:       input.Filename,
		Size:           input.Size,
		EncryptionType: input.EncryptionType,
		StoragePath:    storageKey,
		AccessMode:     input.AccessMode,
		Whitelist:      []string{},
	}

	// Nếu public, sinh CID
	if input.AccessMode == string(models.AccessModePublic) {
		file.PublicCID = uuid.New().String()
	}

	if err := s.fileRepo.Create(ctx, file); err != nil {
		// Rollback: xóa file vừa upload nếu lưu DB thất bại
		_ = s.r2.DeleteFile(ctx, storageKey)
		return nil, fmt.Errorf("failed to save metadata: %w", err)
	}

	if err := s.userStorageRepo.AdjustUsage(ctx, input.OwnerID, input.Size, 1); err != nil {
		_ = s.r2.DeleteFile(ctx, storageKey)
		_ = s.fileRepo.Delete(ctx, file.ID.Hex())
		return nil, fmt.Errorf("failed to update storage usage: %w", err)
	}

	return file, nil
}

// ShareInput là input để cập nhật access mode hoặc whitelist của file
type ShareInput struct {
	FileID      string
	OwnerID     string
	OwnerName   string
	OwnerEmail  string
	OwnerAvatar string
	Filename    string
	AccessMode  string   // "public" | "private" | "whitelist"
	Whitelist   []string // dùng khi AccessMode == "whitelist"
}

// Share cập nhật chế độ chia sẻ / whitelist của file (chỉ owner mới được phép)
func (s *Service) Share(ctx context.Context, input ShareInput) (*models.FileMetadata, error) {
	file, err := s.fileRepo.FindByID(ctx, input.FileID)
	if err != nil {
		return nil, storage.ErrFileNotFound
	}

	if file.OwnerID != input.OwnerID {
		return nil, fmt.Errorf("access denied")
	}

	nextAccessMode := models.AccessMode(input.AccessMode)
	switch nextAccessMode {
	case models.AccessModePublic, models.AccessModePrivate, models.AccessModeWhitelist:
	default:
		return nil, fmt.Errorf("invalid access_mode: %s", input.AccessMode)
	}

	nextFilename := strings.TrimSpace(input.Filename)
	if nextFilename == "" {
		return nil, fmt.Errorf("filename is required")
	}

	oldWhitelist := append([]string(nil), file.Whitelist...)

	file.Filename = nextFilename
	file.AccessMode = string(nextAccessMode)

	if nextAccessMode == models.AccessModePublic && file.PublicCID == "" {
		file.PublicCID = uuid.New().String()
	}

	if nextAccessMode == models.AccessModeWhitelist {
		file.Whitelist = normalizeWhitelist(input.Whitelist, input.OwnerID)
	} else {
		file.Whitelist = []string{}
	}

	if err := s.fileRepo.Update(ctx, file); err != nil {
		return nil, fmt.Errorf("failed to update file: %w", err)
	}

	if s.notificationRepo != nil && nextAccessMode == models.AccessModeWhitelist {
		newlyAddedRecipients := diffWhitelist(oldWhitelist, file.Whitelist)
		if len(newlyAddedRecipients) > 0 {
			notifications := make([]*models.Notification, 0, len(newlyAddedRecipients))
			for _, recipientUserID := range newlyAddedRecipients {
				notifications = append(notifications, &models.Notification{
					RecipientUserID:      recipientUserID,
					ActorUserID:          input.OwnerID,
					Type:                 models.NotificationTypeFileAddedToWhitelist,
					FileID:               file.ID.Hex(),
					FileFilenameSnapshot: file.Filename,
					ActorNameSnapshot:    strings.TrimSpace(input.OwnerName),
					ActorEmailSnapshot:   strings.TrimSpace(input.OwnerEmail),
					ActorAvatarSnapshot:  strings.TrimSpace(input.OwnerAvatar),
				})
			}
			if err := s.notificationRepo.CreateMany(ctx, notifications); err != nil {
				log.Printf("[notifications.warn] failed to create notifications for file=%s err=%v", file.ID.Hex(), err)
			}
		}
	}

	return file, nil
}

func (s *Service) GetDownloadURL(ctx context.Context, fileID, requesterID string) (string, error) {
	file, err := s.fileRepo.FindByID(ctx, fileID)
	if err != nil {
		return "", storage.ErrFileNotFound
	}

	switch models.AccessMode(file.AccessMode) {
	case models.AccessModePublic:
		// Public: ai cũng download được
	case models.AccessModePrivate:
		if file.OwnerID != requesterID {
			return "", fmt.Errorf("access denied")
		}
	case models.AccessModeWhitelist:
		if file.OwnerID != requesterID && !contains(file.Whitelist, requesterID) {
			return "", fmt.Errorf("access denied")
		}
	}

	url, err := s.r2.GeneratePresignedURL(ctx, file.StoragePath, time.Hour)
	if err != nil {
		return "", err
	}

	return url, nil
}

// Delete kiểm tra quyền owner, xóa file trên R2 và MongoDB
func (s *Service) Delete(ctx context.Context, fileID, ownerID string) error {
	file, err := s.fileRepo.FindByID(ctx, fileID)
	if err != nil {
		return storage.ErrFileNotFound
	}

	if file.OwnerID != ownerID {
		return fmt.Errorf("access denied")
	}

	if err := s.r2.DeleteFile(ctx, file.StoragePath); err != nil {
		return err
	}

	if err := s.fileRepo.Delete(ctx, fileID); err != nil {
		return err
	}

	return s.userStorageRepo.AdjustUsage(ctx, ownerID, -file.Size, -1)
}

// ListFiles trả về danh sách file của owner và các file được chia sẻ cho user hiện tại.
func (s *Service) ListFiles(ctx context.Context, userID string) (*FileListResponse, error) {
	ownedFiles, err := s.fileRepo.FindByOwner(ctx, userID)
	if err != nil {
		return nil, err
	}

	sharedFiles, err := s.fileRepo.FindSharedWithUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &FileListResponse{
		OwnedFiles:   ownedFiles,
		SharedWithMe: sharedFiles,
	}, nil
}

func (s *Service) GetMyStorage(ctx context.Context, userID string) (*StorageUsageResponse, error) {
	usage, err := s.userStorageRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if usage.UsedBytes == 0 && usage.FileCount == 0 {
		files, err := s.fileRepo.FindByOwner(ctx, userID)
		if err != nil {
			return nil, err
		}

		if len(files) > 0 {
			var totalBytes int64
			for _, file := range files {
				totalBytes += file.Size
			}
			totalFiles := int64(len(files))

			if err := s.userStorageRepo.SetUsageIfEmpty(ctx, userID, totalBytes, totalFiles); err != nil {
				log.Printf("[storage.debug] failed to backfill usage for user=%s: %v", userID, err)
			}

			usage.UsedBytes = totalBytes
			usage.FileCount = totalFiles
		}
	}

	available := usage.QuotaBytes - usage.UsedBytes
	if available < 0 {
		available = 0
	}

	quotaGB := float64(usage.QuotaBytes) / float64(models.BytesPerGB)
	usedGB := float64(usage.UsedBytes) / float64(models.BytesPerGB)
	percentUsed := 0.0
	if usage.QuotaBytes > 0 {
		percentUsed = (float64(usage.UsedBytes) / float64(usage.QuotaBytes)) * 100
	}
	if percentUsed > 100 {
		percentUsed = 100
	}

	return &StorageUsageResponse{
		UsedBytes:      usage.UsedBytes,
		QuotaBytes:     usage.QuotaBytes,
		UsedGB:         usedGB,
		QuotaGB:        quotaGB,
		PercentUsed:    percentUsed,
		FileCount:      usage.FileCount,
		AvailableBytes: available,
	}, nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func normalizeWhitelist(items []string, ownerID string) []string {
	if len(items) == 0 {
		return []string{}
	}

	seen := make(map[string]struct{}, len(items))
	result := make([]string, 0, len(items))
	for _, item := range items {
		normalized := strings.TrimSpace(item)
		if normalized == "" || normalized == ownerID {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}
	return result
}

func diffWhitelist(previous, current []string) []string {
	if len(current) == 0 {
		return []string{}
	}

	seen := make(map[string]struct{}, len(previous))
	for _, item := range previous {
		seen[item] = struct{}{}
	}

	result := make([]string, 0, len(current))
	for _, item := range current {
		if _, ok := seen[item]; ok {
			continue
		}
		result = append(result, item)
	}
	return result
}

func encryptedSizeFromPlaintextSize(plainSize int64) int64 {
	if plainSize <= 0 {
		return int64(len(crypto.HeaderMagic) + 1 + 16 + 12)
	}

	chunkCount := (plainSize + int64(crypto.ChunkSize) - 1) / int64(crypto.ChunkSize)
	chunkOverhead := chunkCount * (4 + 16)
	headerSize := int64(len(crypto.HeaderMagic) + 1 + 16 + 12)
	return headerSize + plainSize + chunkOverhead
}

// encryptionTypeToBits chuyển chuỗi "AES-128"/"AES-192"/"AES-256" thành crypto.KeyBits
func encryptionTypeToBits(t string) (crypto.KeyBits, error) {
	switch models.EncryptionType(t) {
	case models.EncryptionAES128:
		return crypto.KeyBits128, nil
	case models.EncryptionAES192:
		return crypto.KeyBits192, nil
	case models.EncryptionAES256:
		return crypto.KeyBits256, nil
	default:
		return 0, fmt.Errorf("unsupported encryption type: %s", t)
	}
}
