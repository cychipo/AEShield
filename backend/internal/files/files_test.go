package files

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/aeshield/backend/internal/storage"
	"github.com/aeshield/backend/models"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ---------------------------------------------------------------------------
// Mocks
// ---------------------------------------------------------------------------

type mockR2 struct {
	uploadErr    error
	deleteErr    error
	presignedURL string
	presignedErr error
	uploadedKeys []string
	deletedKeys  []string
}

func (m *mockR2) UploadFile(_ context.Context, key string, _ io.Reader, _ string, size int64) error {
	if m.uploadErr != nil {
		return m.uploadErr
	}
	if size <= 0 {
		return fmt.Errorf("content length must be positive")
	}
	m.uploadedKeys = append(m.uploadedKeys, key)
	return nil
}

func (m *mockR2) DeleteFile(_ context.Context, key string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	m.deletedKeys = append(m.deletedKeys, key)
	return nil
}

func (m *mockR2) GeneratePresignedURL(_ context.Context, _ string, _ time.Duration) (string, error) {
	if m.presignedErr != nil {
		return "", m.presignedErr
	}
	if m.presignedURL == "" {
		return "https://r2.example.com/presigned", nil
	}
	return m.presignedURL, nil
}

// ---------------------------------------------------------------------------

type mockFileRepo struct {
	files     map[string]*models.FileMetadata
	createErr error
	findErr   error
	deleteErr error
}

type mockUserStorageRepo struct {
	storage            map[string]*models.UserStorage
	getErr             error
	adjustErr          error
	setUsageIfEmptyErr error
	adjustLog          []struct {
		userID         string
		usedBytesDelta int64
		fileCountDelta int64
	}
}

func newMockFileRepo() *mockFileRepo {
	return &mockFileRepo{files: make(map[string]*models.FileMetadata)}
}

func newMockUserStorageRepo() *mockUserStorageRepo {
	return &mockUserStorageRepo{storage: make(map[string]*models.UserStorage)}
}

func (m *mockFileRepo) Create(_ context.Context, file *models.FileMetadata) error {
	if m.createErr != nil {
		return m.createErr
	}
	file.ID = primitive.NewObjectID()
	m.files[file.ID.Hex()] = file
	return nil
}

func (m *mockFileRepo) FindByID(_ context.Context, id string) (*models.FileMetadata, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	f, ok := m.files[id]
	if !ok {
		return nil, storage.ErrFileNotFound
	}
	return f, nil
}

func (m *mockFileRepo) FindByOwner(_ context.Context, ownerID string) ([]*models.FileMetadata, error) {
	var result []*models.FileMetadata
	for _, f := range m.files {
		if f.OwnerID == ownerID {
			result = append(result, f)
		}
	}
	return result, nil
}

func (m *mockFileRepo) Update(_ context.Context, file *models.FileMetadata) error {
	m.files[file.ID.Hex()] = file
	return nil
}

func (m *mockFileRepo) Delete(_ context.Context, id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	if _, ok := m.files[id]; !ok {
		return storage.ErrFileNotFound
	}
	delete(m.files, id)
	return nil
}

func (m *mockUserStorageRepo) GetByUserID(_ context.Context, userID string) (*models.UserStorage, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	if st, ok := m.storage[userID]; ok {
		return st, nil
	}
	st := &models.UserStorage{
		UserID:     userID,
		UsedBytes:  0,
		FileCount:  0,
		QuotaBytes: models.DefaultUserQuotaBytes,
	}
	m.storage[userID] = st
	return st, nil
}

func (m *mockUserStorageRepo) AdjustUsage(_ context.Context, userID string, usedBytesDelta, fileCountDelta int64) error {
	m.adjustLog = append(m.adjustLog, struct {
		userID         string
		usedBytesDelta int64
		fileCountDelta int64
	}{
		userID:         userID,
		usedBytesDelta: usedBytesDelta,
		fileCountDelta: fileCountDelta,
	})

	if m.adjustErr != nil {
		return m.adjustErr
	}

	st, ok := m.storage[userID]
	if !ok {
		st = &models.UserStorage{UserID: userID, QuotaBytes: models.DefaultUserQuotaBytes}
		m.storage[userID] = st
	}
	st.UsedBytes += usedBytesDelta
	st.FileCount += fileCountDelta
	return nil
}

func (m *mockUserStorageRepo) SetUsageIfEmpty(_ context.Context, userID string, usedBytes, fileCount int64) error {
	if m.setUsageIfEmptyErr != nil {
		return m.setUsageIfEmptyErr
	}

	st, ok := m.storage[userID]
	if !ok {
		st = &models.UserStorage{UserID: userID, QuotaBytes: models.DefaultUserQuotaBytes}
		m.storage[userID] = st
	}

	if st.UsedBytes == 0 && st.FileCount == 0 {
		st.UsedBytes = usedBytes
		st.FileCount = fileCount
	}

	return nil
}

type mockHandlerService struct {
	uploadErr  error
	storage    *StorageUsageResponse
	storageErr error
}

func (m *mockHandlerService) Upload(_ context.Context, _ UploadInput) (*models.FileMetadata, error) {
	if m.uploadErr != nil {
		return nil, m.uploadErr
	}
	return &models.FileMetadata{}, nil
}

func (m *mockHandlerService) GetDownloadURL(_ context.Context, _ string, _ string) (string, error) {
	return "https://r2.example.com/presigned", nil
}

func (m *mockHandlerService) Delete(_ context.Context, _ string, _ string) error {
	return nil
}

func (m *mockHandlerService) ListFiles(_ context.Context, _ string) ([]*models.FileMetadata, error) {
	return []*models.FileMetadata{}, nil
}

func (m *mockHandlerService) Share(_ context.Context, _ ShareInput) (*models.FileMetadata, error) {
	return &models.FileMetadata{}, nil
}

func (m *mockHandlerService) GetMyStorage(_ context.Context, _ string) (*StorageUsageResponse, error) {
	if m.storageErr != nil {
		return nil, m.storageErr
	}
	if m.storage != nil {
		return m.storage, nil
	}
	return &StorageUsageResponse{}, nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func newTestService(r2 *mockR2, repo *mockFileRepo, userStorageRepo *mockUserStorageRepo) *Service {
	return NewService(r2, repo, userStorageRepo)
}

func seedFile(repo *mockFileRepo, ownerID, accessMode string, whitelist []string) *models.FileMetadata {
	f := &models.FileMetadata{
		ID:          primitive.NewObjectID(),
		OwnerID:     ownerID,
		Filename:    "test.txt",
		Size:        100,
		StoragePath: "owner/uuid.txt",
		AccessMode:  accessMode,
		Whitelist:   whitelist,
	}
	repo.files[f.ID.Hex()] = f
	return f
}

// ---------------------------------------------------------------------------
// Service tests — Upload
// ---------------------------------------------------------------------------

func TestUpload_Success_Private(t *testing.T) {
	r2 := &mockR2{}
	repo := newMockFileRepo()
	svc := newTestService(r2, repo, newMockUserStorageRepo())

	file, err := svc.Upload(context.Background(), UploadInput{
		OwnerID:        "user-1",
		Filename:       "hello.txt",
		Size:           42,
		ContentType:    "text/plain",
		EncryptionType: "AES-256",
		AccessMode:     "private",
		Password:       "secret",
		Body:           strings.NewReader("hello world"),
	})

	require.NoError(t, err)
	assert.Equal(t, "user-1", file.OwnerID)
	assert.Equal(t, "hello.txt", file.Filename)
	assert.Equal(t, "private", file.AccessMode)
	assert.Empty(t, file.PublicCID) // private không có CID
	assert.Len(t, r2.uploadedKeys, 1)
	assert.Contains(t, r2.uploadedKeys[0], "user-1/")
}

func TestUpload_Success_Public_GeneratesCID(t *testing.T) {
	r2 := &mockR2{}
	repo := newMockFileRepo()
	svc := newTestService(r2, repo, newMockUserStorageRepo())

	file, err := svc.Upload(context.Background(), UploadInput{
		OwnerID:        "user-1",
		Filename:       "photo.png",
		Size:           1024,
		AccessMode:     "public",
		EncryptionType: "AES-256",
		Password:       "secret",
		Body:           bytes.NewReader([]byte("image data")),
	})

	require.NoError(t, err)
	assert.Equal(t, "public", file.AccessMode)
	assert.NotEmpty(t, file.PublicCID)
}

func TestUpload_ContentType_DetectedFromExtension(t *testing.T) {
	r2 := &mockR2{}
	repo := newMockFileRepo()
	svc := newTestService(r2, repo, newMockUserStorageRepo())

	_, err := svc.Upload(context.Background(), UploadInput{
		OwnerID:        "user-1",
		Filename:       "doc.pdf",
		Size:           500,
		EncryptionType: "AES-256",
		Password:       "secret",
		Body:           strings.NewReader("pdf content"),
		// ContentType để trống → tự detect từ .pdf
	})

	require.NoError(t, err)
}

func TestUpload_R2Fails_ReturnsError(t *testing.T) {
	r2 := &mockR2{uploadErr: errors.New("R2 unavailable")}
	repo := newMockFileRepo()
	svc := newTestService(r2, repo, newMockUserStorageRepo())

	_, err := svc.Upload(context.Background(), UploadInput{
		OwnerID:        "user-1",
		Filename:       "file.txt",
		EncryptionType: "AES-256",
		Password:       "secret",
		Body:           strings.NewReader("data"),
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "upload failed")
}

func TestUpload_DBFails_RollbackR2(t *testing.T) {
	r2 := &mockR2{}
	repo := newMockFileRepo()
	repo.createErr = errors.New("mongo timeout")
	svc := newTestService(r2, repo, newMockUserStorageRepo())

	_, err := svc.Upload(context.Background(), UploadInput{
		OwnerID:        "user-1",
		Filename:       "file.txt",
		Size:           4,
		EncryptionType: "AES-256",
		Password:       "secret",
		Body:           strings.NewReader("data"),
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to save metadata")
	// Phải rollback xóa file trên R2
	assert.Len(t, r2.deletedKeys, 1)
}

func TestEncryptedSizeFromPlaintextSize(t *testing.T) {
	assert.Equal(t, int64(33), encryptedSizeFromPlaintextSize(0))
	assert.Equal(t, int64(54), encryptedSizeFromPlaintextSize(1))
	assert.Equal(t, int64(65589), encryptedSizeFromPlaintextSize(65536))
	assert.Equal(t, int64(65610), encryptedSizeFromPlaintextSize(65537))
}

func TestUpload_SizeZero_FallbackBufferStillUploads(t *testing.T) {
	r2 := &mockR2{}
	repo := newMockFileRepo()
	svc := newTestService(r2, repo, newMockUserStorageRepo())

	_, err := svc.Upload(context.Background(), UploadInput{
		OwnerID:        "user-1",
		Filename:       "unknown.bin",
		Size:           0,
		EncryptionType: "AES-256",
		Password:       "secret",
		Body:           strings.NewReader("payload"),
	})

	require.NoError(t, err)
	assert.Len(t, r2.uploadedKeys, 1)
}

func TestUpload_FileTooLarge_ReturnsDomainError(t *testing.T) {
	r2 := &mockR2{}
	repo := newMockFileRepo()
	userStorageRepo := newMockUserStorageRepo()
	svc := newTestService(r2, repo, userStorageRepo)

	_, err := svc.Upload(context.Background(), UploadInput{
		OwnerID:        "user-1",
		Filename:       "large.bin",
		Size:           models.MaxFileSizeBytes + 1,
		EncryptionType: "AES-256",
		Password:       "secret",
		Body:           strings.NewReader("payload"),
	})

	require.Error(t, err)
	assert.Equal(t, ErrFileTooLarge, err)
	assert.Empty(t, r2.uploadedKeys)
}

func TestUpload_QuotaExceeded_ReturnsDomainError(t *testing.T) {
	r2 := &mockR2{}
	repo := newMockFileRepo()
	userStorageRepo := newMockUserStorageRepo()
	userStorageRepo.storage["user-1"] = &models.UserStorage{
		UserID:     "user-1",
		UsedBytes:  models.DefaultUserQuotaBytes - 10,
		FileCount:  5,
		QuotaBytes: models.DefaultUserQuotaBytes,
	}
	svc := newTestService(r2, repo, userStorageRepo)

	_, err := svc.Upload(context.Background(), UploadInput{
		OwnerID:        "user-1",
		Filename:       "over-quota.txt",
		Size:           20,
		EncryptionType: "AES-256",
		Password:       "secret",
		Body:           strings.NewReader("data"),
	})

	require.Error(t, err)
	assert.Equal(t, ErrStorageQuota, err)
	assert.Empty(t, r2.uploadedKeys)
}

func TestUpload_AdjustUsageFails_RollbackMetadataAndR2(t *testing.T) {
	r2 := &mockR2{}
	repo := newMockFileRepo()
	userStorageRepo := newMockUserStorageRepo()
	userStorageRepo.adjustErr = errors.New("adjust failed")
	svc := newTestService(r2, repo, userStorageRepo)

	file, err := svc.Upload(context.Background(), UploadInput{
		OwnerID:        "user-1",
		Filename:       "file.txt",
		Size:           4,
		EncryptionType: "AES-256",
		Password:       "secret",
		Body:           strings.NewReader("data"),
	})

	require.Error(t, err)
	assert.Nil(t, file)
	assert.Contains(t, err.Error(), "failed to update storage usage")
	assert.Len(t, r2.deletedKeys, 1)
	assert.Empty(t, repo.files)
}

func TestUpload_AndDelete_AdjustUsageDeltas(t *testing.T) {
	r2 := &mockR2{}
	repo := newMockFileRepo()
	userStorageRepo := newMockUserStorageRepo()
	svc := newTestService(r2, repo, userStorageRepo)

	uploaded, err := svc.Upload(context.Background(), UploadInput{
		OwnerID:        "user-1",
		Filename:       "track.txt",
		Size:           128,
		EncryptionType: "AES-256",
		AccessMode:     "private",
		Password:       "secret",
		Body:           strings.NewReader("track-data"),
	})
	require.NoError(t, err)
	require.NotNil(t, uploaded)

	err = svc.Delete(context.Background(), uploaded.ID.Hex(), "user-1")
	require.NoError(t, err)
	require.Len(t, userStorageRepo.adjustLog, 2)

	assert.Equal(t, int64(128), userStorageRepo.adjustLog[0].usedBytesDelta)
	assert.Equal(t, int64(1), userStorageRepo.adjustLog[0].fileCountDelta)
	assert.Equal(t, int64(-128), userStorageRepo.adjustLog[1].usedBytesDelta)
	assert.Equal(t, int64(-1), userStorageRepo.adjustLog[1].fileCountDelta)
}

func TestGetMyStorage_ReturnsCalculatedSchema(t *testing.T) {
	r2 := &mockR2{}
	repo := newMockFileRepo()
	userStorageRepo := newMockUserStorageRepo()
	userStorageRepo.storage["user-1"] = &models.UserStorage{
		UserID:     "user-1",
		UsedBytes:  2 * models.BytesPerGB,
		FileCount:  12,
		QuotaBytes: 10 * models.BytesPerGB,
	}
	svc := newTestService(r2, repo, userStorageRepo)

	usage, err := svc.GetMyStorage(context.Background(), "user-1")
	require.NoError(t, err)
	require.NotNil(t, usage)

	assert.Equal(t, int64(2*models.BytesPerGB), usage.UsedBytes)
	assert.Equal(t, int64(10*models.BytesPerGB), usage.QuotaBytes)
	assert.Equal(t, 2.0, usage.UsedGB)
	assert.Equal(t, 10.0, usage.QuotaGB)
	assert.Equal(t, 20.0, usage.PercentUsed)
	assert.Equal(t, int64(12), usage.FileCount)
	assert.Equal(t, int64(8*models.BytesPerGB), usage.AvailableBytes)
}

func TestGetMyStorage_BackfillsFromExistingFiles(t *testing.T) {
	r2 := &mockR2{}
	repo := newMockFileRepo()
	repo.files[primitive.NewObjectID().Hex()] = &models.FileMetadata{OwnerID: "user-1", Size: 2048}
	repo.files[primitive.NewObjectID().Hex()] = &models.FileMetadata{OwnerID: "user-1", Size: 1024}
	repo.files[primitive.NewObjectID().Hex()] = &models.FileMetadata{OwnerID: "user-2", Size: 999}

	userStorageRepo := newMockUserStorageRepo()
	userStorageRepo.storage["user-1"] = &models.UserStorage{
		UserID:     "user-1",
		UsedBytes:  0,
		FileCount:  0,
		QuotaBytes: models.DefaultUserQuotaBytes,
	}

	svc := newTestService(r2, repo, userStorageRepo)
	usage, err := svc.GetMyStorage(context.Background(), "user-1")
	require.NoError(t, err)
	require.NotNil(t, usage)

	assert.Equal(t, int64(3072), usage.UsedBytes)
	assert.Equal(t, int64(2), usage.FileCount)
	assert.Equal(t, int64(3072), userStorageRepo.storage["user-1"].UsedBytes)
	assert.Equal(t, int64(2), userStorageRepo.storage["user-1"].FileCount)
}

// ---------------------------------------------------------------------------
// Service tests — GetDownloadURL
// ---------------------------------------------------------------------------

func TestGetDownloadURL_PublicFile_AnyoneCanDownload(t *testing.T) {
	r2 := &mockR2{}
	repo := newMockFileRepo()
	f := seedFile(repo, "owner-1", "public", nil)
	svc := newTestService(r2, repo, newMockUserStorageRepo())

	url, err := svc.GetDownloadURL(context.Background(), f.ID.Hex(), "stranger")

	require.NoError(t, err)
	assert.NotEmpty(t, url)
}

func TestGetDownloadURL_PrivateFile_OwnerCanDownload(t *testing.T) {
	r2 := &mockR2{}
	repo := newMockFileRepo()
	f := seedFile(repo, "owner-1", "private", nil)
	svc := newTestService(r2, repo, newMockUserStorageRepo())

	url, err := svc.GetDownloadURL(context.Background(), f.ID.Hex(), "owner-1")

	require.NoError(t, err)
	assert.NotEmpty(t, url)
}

func TestGetDownloadURL_PrivateFile_StrangerDenied(t *testing.T) {
	r2 := &mockR2{}
	repo := newMockFileRepo()
	f := seedFile(repo, "owner-1", "private", nil)
	svc := newTestService(r2, repo, newMockUserStorageRepo())

	_, err := svc.GetDownloadURL(context.Background(), f.ID.Hex(), "stranger")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "access denied")
}

func TestGetDownloadURL_WhitelistFile_AllowedUserCanDownload(t *testing.T) {
	r2 := &mockR2{}
	repo := newMockFileRepo()
	f := seedFile(repo, "owner-1", "whitelist", []string{"allowed-user"})
	svc := newTestService(r2, repo, newMockUserStorageRepo())

	url, err := svc.GetDownloadURL(context.Background(), f.ID.Hex(), "allowed-user")

	require.NoError(t, err)
	assert.NotEmpty(t, url)
}

func TestGetDownloadURL_WhitelistFile_UnauthorizedDenied(t *testing.T) {
	r2 := &mockR2{}
	repo := newMockFileRepo()
	f := seedFile(repo, "owner-1", "whitelist", []string{"allowed-user"})
	svc := newTestService(r2, repo, newMockUserStorageRepo())

	_, err := svc.GetDownloadURL(context.Background(), f.ID.Hex(), "random-user")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "access denied")
}

func TestGetDownloadURL_FileNotFound(t *testing.T) {
	r2 := &mockR2{}
	repo := newMockFileRepo()
	svc := newTestService(r2, repo, newMockUserStorageRepo())

	_, err := svc.GetDownloadURL(context.Background(), primitive.NewObjectID().Hex(), "user-1")

	require.Error(t, err)
	assert.Equal(t, storage.ErrFileNotFound, err)
}

func TestGetDownloadURL_PresignFails(t *testing.T) {
	r2 := &mockR2{presignedErr: errors.New("presign error")}
	repo := newMockFileRepo()
	f := seedFile(repo, "owner-1", "public", nil)
	svc := newTestService(r2, repo, newMockUserStorageRepo())

	_, err := svc.GetDownloadURL(context.Background(), f.ID.Hex(), "anyone")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "presign error")
}

// ---------------------------------------------------------------------------
// Service tests — Delete
// ---------------------------------------------------------------------------

func TestDelete_OwnerCanDelete(t *testing.T) {
	r2 := &mockR2{}
	repo := newMockFileRepo()
	f := seedFile(repo, "owner-1", "private", nil)
	svc := newTestService(r2, repo, newMockUserStorageRepo())

	err := svc.Delete(context.Background(), f.ID.Hex(), "owner-1")

	require.NoError(t, err)
	assert.Len(t, r2.deletedKeys, 1)
	assert.Equal(t, f.StoragePath, r2.deletedKeys[0])
	assert.Empty(t, repo.files)
}

func TestDelete_NonOwnerDenied(t *testing.T) {
	r2 := &mockR2{}
	repo := newMockFileRepo()
	f := seedFile(repo, "owner-1", "private", nil)
	svc := newTestService(r2, repo, newMockUserStorageRepo())

	err := svc.Delete(context.Background(), f.ID.Hex(), "other-user")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "access denied")
	assert.Empty(t, r2.deletedKeys) // R2 không bị xóa
}

func TestDelete_FileNotFound(t *testing.T) {
	r2 := &mockR2{}
	repo := newMockFileRepo()
	svc := newTestService(r2, repo, newMockUserStorageRepo())

	err := svc.Delete(context.Background(), primitive.NewObjectID().Hex(), "owner-1")

	require.Error(t, err)
	assert.Equal(t, storage.ErrFileNotFound, err)
}

// ---------------------------------------------------------------------------
// Service tests — ListFiles
// ---------------------------------------------------------------------------

func TestListFiles_ReturnsOnlyOwnerFiles(t *testing.T) {
	r2 := &mockR2{}
	repo := newMockFileRepo()
	seedFile(repo, "user-A", "private", nil)
	seedFile(repo, "user-A", "public", nil)
	seedFile(repo, "user-B", "private", nil)
	svc := newTestService(r2, repo, newMockUserStorageRepo())

	files, err := svc.ListFiles(context.Background(), "user-A")

	require.NoError(t, err)
	assert.Len(t, files, 2)
	for _, f := range files {
		assert.Equal(t, "user-A", f.OwnerID)
	}
}

func TestListFiles_Empty(t *testing.T) {
	r2 := &mockR2{}
	repo := newMockFileRepo()
	svc := newTestService(r2, repo, newMockUserStorageRepo())

	files, err := svc.ListFiles(context.Background(), "user-nobody")

	require.NoError(t, err)
	assert.Empty(t, files)
}

// ---------------------------------------------------------------------------
// Handler tests (HTTP layer)
// ---------------------------------------------------------------------------

func setupHandlerApp(r2 *mockR2, repo *mockFileRepo, userStorageRepo *mockUserStorageRepo, userID string) *fiber.App {
	svc := newTestService(r2, repo, userStorageRepo)
	h := NewHandler(svc)

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		},
	})

	// Inject fake JWT claims
	app.Use(func(c *fiber.Ctx) error {
		if userID != "" {
			c.Locals("user", &models.Claims{UserID: userID, Email: userID + "@test.com"})
		}
		return c.Next()
	})

	app.Post("/files/upload", h.Upload)
	app.Get("/files", h.ListFiles)
	app.Get("/files/:id/download", h.Download)
	app.Delete("/files/:id", h.Delete)
	app.Patch("/files/share", h.Share)
	app.Patch("/files/:id", h.Share)
	app.Get("/storage/me", h.GetMyStorage)

	return app
}

func TestHandler_Upload_Success(t *testing.T) {
	r2 := &mockR2{}
	repo := newMockFileRepo()
	app := setupHandlerApp(r2, repo, newMockUserStorageRepo(), "user-1")

	body := &bytes.Buffer{}
	body.WriteString("--boundary\r\nContent-Disposition: form-data; name=\"file\"; filename=\"test.txt\"\r\nContent-Type: text/plain\r\n\r\nhello\r\n--boundary\r\nContent-Disposition: form-data; name=\"password\"\r\n\r\nsecret\r\n--boundary--\r\n")

	req := httptest.NewRequest("POST", "/files/upload", body)
	req.Header.Set("Content-Type", "multipart/form-data; boundary=boundary")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)
}

func TestHandler_Upload_Unauthorized(t *testing.T) {
	r2 := &mockR2{}
	repo := newMockFileRepo()
	app := setupHandlerApp(r2, repo, newMockUserStorageRepo(), "") // no user

	req := httptest.NewRequest("POST", "/files/upload", nil)
	req.Header.Set("Content-Type", "multipart/form-data; boundary=boundary")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestHandler_Upload_InvalidEncryptionType(t *testing.T) {
	r2 := &mockR2{}
	repo := newMockFileRepo()
	app := setupHandlerApp(r2, repo, newMockUserStorageRepo(), "user-1")

	body := &bytes.Buffer{}
	body.WriteString("--boundary\r\nContent-Disposition: form-data; name=\"file\"; filename=\"test.txt\"\r\nContent-Type: text/plain\r\n\r\nhello\r\n--boundary\r\nContent-Disposition: form-data; name=\"password\"\r\n\r\nsecret\r\n--boundary\r\nContent-Disposition: form-data; name=\"encryption_type\"\r\n\r\nAES-999\r\n--boundary--\r\n")

	req := httptest.NewRequest("POST", "/files/upload", body)
	req.Header.Set("Content-Type", "multipart/form-data; boundary=boundary")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandler_Upload_QuotaExceeded_Returns413(t *testing.T) {
	r2 := &mockR2{}
	repo := newMockFileRepo()
	userStorageRepo := newMockUserStorageRepo()
	userStorageRepo.storage["user-1"] = &models.UserStorage{
		UserID:     "user-1",
		UsedBytes:  models.DefaultUserQuotaBytes,
		FileCount:  2,
		QuotaBytes: models.DefaultUserQuotaBytes,
	}
	app := setupHandlerApp(r2, repo, userStorageRepo, "user-1")

	body := &bytes.Buffer{}
	body.WriteString("--boundary\r\nContent-Disposition: form-data; name=\"file\"; filename=\"small.txt\"\r\nContent-Type: text/plain\r\n\r\nhello\r\n--boundary\r\nContent-Disposition: form-data; name=\"password\"\r\n\r\nsecret\r\n--boundary--\r\n")

	req := httptest.NewRequest("POST", "/files/upload", body)
	req.Header.Set("Content-Type", "multipart/form-data; boundary=boundary")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusRequestEntityTooLarge, resp.StatusCode)
}

func TestHandler_Upload_FileTooLarge_Returns413(t *testing.T) {
	h := NewHandler(&mockHandlerService{uploadErr: ErrFileTooLarge})
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user", &models.Claims{UserID: "user-1", Email: "u@test.com"})
		return c.Next()
	})
	app.Post("/files/upload", h.Upload)

	body := &bytes.Buffer{}
	body.WriteString("--boundary\r\nContent-Disposition: form-data; name=\"file\"; filename=\"test.txt\"\r\nContent-Type: text/plain\r\n\r\nhello\r\n--boundary\r\nContent-Disposition: form-data; name=\"password\"\r\n\r\nsecret\r\n--boundary--\r\n")

	req := httptest.NewRequest("POST", "/files/upload", body)
	req.Header.Set("Content-Type", "multipart/form-data; boundary=boundary")
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusRequestEntityTooLarge, resp.StatusCode)
}

func TestHandler_GetMyStorage_Success(t *testing.T) {
	r2 := &mockR2{}
	repo := newMockFileRepo()
	userStorageRepo := newMockUserStorageRepo()
	userStorageRepo.storage["user-1"] = &models.UserStorage{
		UserID:     "user-1",
		UsedBytes:  3 * models.BytesPerGB,
		FileCount:  9,
		QuotaBytes: 10 * models.BytesPerGB,
	}
	app := setupHandlerApp(r2, repo, userStorageRepo, "user-1")

	req := httptest.NewRequest("GET", "/storage/me", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var usage StorageUsageResponse
	err = json.NewDecoder(resp.Body).Decode(&usage)
	require.NoError(t, err)
	assert.Equal(t, int64(3*models.BytesPerGB), usage.UsedBytes)
	assert.Equal(t, int64(10*models.BytesPerGB), usage.QuotaBytes)
	assert.Equal(t, 30.0, usage.PercentUsed)
	assert.Equal(t, int64(9), usage.FileCount)
	assert.Equal(t, int64(7*models.BytesPerGB), usage.AvailableBytes)
}

func TestHandler_ListFiles_Success(t *testing.T) {
	r2 := &mockR2{}
	repo := newMockFileRepo()
	seedFile(repo, "user-1", "private", nil)
	seedFile(repo, "user-1", "public", nil)
	app := setupHandlerApp(r2, repo, newMockUserStorageRepo(), "user-1")

	req := httptest.NewRequest("GET", "/files", nil)
	resp, err := app.Test(req)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var result []models.FileMetadata
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Len(t, result, 2)
}

func TestHandler_Download_Success(t *testing.T) {
	r2 := &mockR2{presignedURL: "https://r2.example.com/file"}
	repo := newMockFileRepo()
	f := seedFile(repo, "user-1", "private", nil)
	app := setupHandlerApp(r2, repo, newMockUserStorageRepo(), "user-1")

	req := httptest.NewRequest("GET", "/files/"+f.ID.Hex()+"/download", nil)
	resp, err := app.Test(req)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var result map[string]string
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, "https://r2.example.com/file", result["url"])
}

func TestHandler_Download_NotFound(t *testing.T) {
	r2 := &mockR2{}
	repo := newMockFileRepo()
	app := setupHandlerApp(r2, repo, newMockUserStorageRepo(), "user-1")

	req := httptest.NewRequest("GET", "/files/"+primitive.NewObjectID().Hex()+"/download", nil)
	resp, err := app.Test(req)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

func TestHandler_Download_Forbidden(t *testing.T) {
	r2 := &mockR2{}
	repo := newMockFileRepo()
	f := seedFile(repo, "owner-1", "private", nil)
	app := setupHandlerApp(r2, repo, newMockUserStorageRepo(), "other-user")

	req := httptest.NewRequest("GET", "/files/"+f.ID.Hex()+"/download", nil)
	resp, err := app.Test(req)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestHandler_Delete_Success(t *testing.T) {
	r2 := &mockR2{}
	repo := newMockFileRepo()
	f := seedFile(repo, "user-1", "private", nil)
	app := setupHandlerApp(r2, repo, newMockUserStorageRepo(), "user-1")

	req := httptest.NewRequest("DELETE", "/files/"+f.ID.Hex(), nil)
	resp, err := app.Test(req)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.Empty(t, repo.files)
}

func TestHandler_Delete_Forbidden(t *testing.T) {
	r2 := &mockR2{}
	repo := newMockFileRepo()
	f := seedFile(repo, "owner-1", "private", nil)
	app := setupHandlerApp(r2, repo, newMockUserStorageRepo(), "other-user")

	req := httptest.NewRequest("DELETE", "/files/"+f.ID.Hex(), nil)
	resp, err := app.Test(req)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestHandler_Delete_NotFound(t *testing.T) {
	r2 := &mockR2{}
	repo := newMockFileRepo()
	app := setupHandlerApp(r2, repo, newMockUserStorageRepo(), "user-1")

	req := httptest.NewRequest("DELETE", "/files/"+primitive.NewObjectID().Hex(), nil)
	resp, err := app.Test(req)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// Service tests — Share
// ---------------------------------------------------------------------------

func TestShare_OwnerCanMakePublic(t *testing.T) {
	r2 := &mockR2{}
	repo := newMockFileRepo()
	f := seedFile(repo, "owner-1", "private", nil)
	svc := newTestService(r2, repo, newMockUserStorageRepo())

	updated, err := svc.Share(context.Background(), ShareInput{
		FileID:     f.ID.Hex(),
		OwnerID:    "owner-1",
		Filename:   "public-file.txt",
		AccessMode: "public",
	})

	require.NoError(t, err)
	assert.Equal(t, "public", updated.AccessMode)
	assert.NotEmpty(t, updated.PublicCID)
}

func TestShare_OwnerCanSetWhitelist(t *testing.T) {
	r2 := &mockR2{}
	repo := newMockFileRepo()
	f := seedFile(repo, "owner-1", "private", nil)
	svc := newTestService(r2, repo, newMockUserStorageRepo())

	updated, err := svc.Share(context.Background(), ShareInput{
		FileID:     f.ID.Hex(),
		OwnerID:    "owner-1",
		Filename:   "whitelist-file.txt",
		AccessMode: "whitelist",
		Whitelist:  []string{"alice@test.com", "bob@test.com"},
	})

	require.NoError(t, err)
	assert.Equal(t, "whitelist", updated.AccessMode)
	assert.ElementsMatch(t, []string{"alice@test.com", "bob@test.com"}, updated.Whitelist)
}

func TestShare_NonOwnerDenied(t *testing.T) {
	r2 := &mockR2{}
	repo := newMockFileRepo()
	f := seedFile(repo, "owner-1", "private", nil)
	svc := newTestService(r2, repo, newMockUserStorageRepo())

	_, err := svc.Share(context.Background(), ShareInput{
		FileID:     f.ID.Hex(),
		OwnerID:    "other-user",
		Filename:   "no-access.txt",
		AccessMode: "public",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "access denied")
}

func TestShare_FileNotFound(t *testing.T) {
	r2 := &mockR2{}
	repo := newMockFileRepo()
	svc := newTestService(r2, repo, newMockUserStorageRepo())

	_, err := svc.Share(context.Background(), ShareInput{
		FileID:     primitive.NewObjectID().Hex(),
		OwnerID:    "owner-1",
		Filename:   "missing-file.txt",
		AccessMode: "public",
	})

	require.Error(t, err)
	assert.Equal(t, storage.ErrFileNotFound, err)
}

func TestShare_InvalidAccessMode(t *testing.T) {
	r2 := &mockR2{}
	repo := newMockFileRepo()
	f := seedFile(repo, "owner-1", "private", nil)
	svc := newTestService(r2, repo, newMockUserStorageRepo())

	_, err := svc.Share(context.Background(), ShareInput{
		FileID:     f.ID.Hex(),
		OwnerID:    "owner-1",
		Filename:   "invalid-mode.txt",
		AccessMode: "badmode",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid access_mode")
}

func TestShare_RequireFilename(t *testing.T) {
	r2 := &mockR2{}
	repo := newMockFileRepo()
	f := seedFile(repo, "owner-1", "private", nil)
	svc := newTestService(r2, repo, newMockUserStorageRepo())

	_, err := svc.Share(context.Background(), ShareInput{
		FileID:     f.ID.Hex(),
		OwnerID:    "owner-1",
		Filename:   "   ",
		AccessMode: "private",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "filename is required")
}

func TestShare_PrivateClearsWhitelist(t *testing.T) {
	r2 := &mockR2{}
	repo := newMockFileRepo()
	f := seedFile(repo, "owner-1", "whitelist", []string{"alice@test.com"})
	svc := newTestService(r2, repo, newMockUserStorageRepo())

	updated, err := svc.Share(context.Background(), ShareInput{
		FileID:     f.ID.Hex(),
		OwnerID:    "owner-1",
		Filename:   "private-file.txt",
		AccessMode: "private",
	})

	require.NoError(t, err)
	assert.Equal(t, "private", updated.AccessMode)
	assert.Empty(t, updated.Whitelist)
	assert.Equal(t, "private-file.txt", updated.Filename)
}

// Handler tests — Share
func TestHandler_Share_Success(t *testing.T) {
	r2 := &mockR2{}
	repo := newMockFileRepo()
	f := seedFile(repo, "user-1", "private", nil)
	app := setupHandlerApp(r2, repo, newMockUserStorageRepo(), "user-1")

	body := fmt.Sprintf(`{"file_id":"%s","filename":"updated-name.txt","access_mode":"public"}`, f.ID.Hex())
	req := httptest.NewRequest("PATCH", "/files/share", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestHandler_Share_MissingFileID(t *testing.T) {
	r2 := &mockR2{}
	repo := newMockFileRepo()
	app := setupHandlerApp(r2, repo, newMockUserStorageRepo(), "user-1")

	req := httptest.NewRequest("PATCH", "/files/share", strings.NewReader(`{"filename":"newname.txt","access_mode":"public"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandler_Share_MissingFilename(t *testing.T) {
	r2 := &mockR2{}
	repo := newMockFileRepo()
	f := seedFile(repo, "user-1", "private", nil)
	app := setupHandlerApp(r2, repo, newMockUserStorageRepo(), "user-1")

	body := fmt.Sprintf(`{"file_id":"%s","access_mode":"public"}`, f.ID.Hex())
	req := httptest.NewRequest("PATCH", "/files/share", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandler_Share_ByPathID_Success(t *testing.T) {
	r2 := &mockR2{}
	repo := newMockFileRepo()
	f := seedFile(repo, "user-1", "private", nil)
	app := setupHandlerApp(r2, repo, newMockUserStorageRepo(), "user-1")

	body := `{"filename":"rename-by-path.txt","access_mode":"private"}`
	req := httptest.NewRequest("PATCH", "/files/"+f.ID.Hex(), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

// Đảm bảo interfaces được implement đúng (compile-time check)
var _ R2Storage = (*mockR2)(nil)
var _ FileRepo = (*mockFileRepo)(nil)
var _ UserStorageRepo = (*mockUserStorageRepo)(nil)
var _ FileService = (*mockHandlerService)(nil)

// Unused import guard
var _ = fmt.Sprintf
