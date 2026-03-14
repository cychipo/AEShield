package files

import (
	"context"
	"fmt"
	"io"
	"mime"
	"path/filepath"
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
	GeneratePresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error)
}

// FileRepo là interface để có thể mock trong tests
type FileRepo interface {
	Create(ctx context.Context, file *models.FileMetadata) error
	FindByID(ctx context.Context, id string) (*models.FileMetadata, error)
	FindByOwner(ctx context.Context, ownerID string) ([]*models.FileMetadata, error)
	Update(ctx context.Context, file *models.FileMetadata) error
	Delete(ctx context.Context, id string) error
}

type Service struct {
	r2       R2Storage
	fileRepo FileRepo
}

func NewService(r2 R2Storage, fileRepo FileRepo) *Service {
	return &Service{
		r2:       r2,
		fileRepo: fileRepo,
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
}

// Upload stream mã hóa file và đẩy lên R2, lưu metadata vào MongoDB.
func (s *Service) Upload(ctx context.Context, input UploadInput) (*models.FileMetadata, error) {
	// Sinh unique key cho R2
	ext := filepath.Ext(input.Filename)
	storageKey := fmt.Sprintf("%s/%s%s", input.OwnerID, uuid.New().String(), ext)

	// Detect content type nếu chưa có
	contentType := input.ContentType
	if contentType == "" || contentType == "application/octet-stream" {
		contentType = mime.TypeByExtension(ext)
		if contentType == "" {
			contentType = "application/octet-stream"
		}
	}

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
		_, encryptErr = crypto.Encrypt(pw, input.Body, input.Password, bits)
	}()

	// Upload từ pipe reader lên R2 (size = -1 vì ciphertext lớn hơn plaintext)
	if err := s.r2.UploadFile(ctx, storageKey, pr, "application/octet-stream", -1); err != nil {
		pr.CloseWithError(err)
		return nil, fmt.Errorf("upload failed: %w", err)
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

	return file, nil
}

// ShareInput là input để cập nhật access mode hoặc whitelist của file
type ShareInput struct {
	FileID     string
	OwnerID    string
	AccessMode string   // "public" | "private" | "whitelist"
	Whitelist  []string // dùng khi AccessMode == "whitelist"
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

	switch models.AccessMode(input.AccessMode) {
	case models.AccessModePublic, models.AccessModePrivate, models.AccessModeWhitelist:
	default:
		return nil, fmt.Errorf("invalid access_mode: %s", input.AccessMode)
	}

	file.AccessMode = input.AccessMode

	if models.AccessMode(input.AccessMode) == models.AccessModePublic && file.PublicCID == "" {
		file.PublicCID = uuid.New().String()
	}

	if models.AccessMode(input.AccessMode) == models.AccessModeWhitelist {
		if input.Whitelist == nil {
			input.Whitelist = []string{}
		}
		file.Whitelist = input.Whitelist
	}

	if err := s.fileRepo.Update(ctx, file); err != nil {
		return nil, fmt.Errorf("failed to update file: %w", err)
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

	return s.fileRepo.Delete(ctx, fileID)
}

// ListFiles trả về danh sách file của owner
func (s *Service) ListFiles(ctx context.Context, ownerID string) ([]*models.FileMetadata, error) {
	return s.fileRepo.FindByOwner(ctx, ownerID)
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
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
