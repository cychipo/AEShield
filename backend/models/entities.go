package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type LinkedProvider struct {
	Provider   string `bson:"provider" json:"provider"`
	ProviderID string `bson:"provider_id" json:"provider_id"`
}

type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Email     string             `bson:"email" json:"email"`
	Avatar    string             `bson:"avatar" json:"avatar"`
	Name      string             `bson:"name" json:"name"`
	Providers []LinkedProvider   `bson:"providers" json:"providers"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

func (User) CollectionName() string {
	return "users"
}

type FileMetadata struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	OwnerID        string             `bson:"owner_id" json:"owner_id"`
	Filename       string             `bson:"filename" json:"filename"`
	Size           int64              `bson:"size" json:"size"`
	EncryptionType string             `bson:"encryption_type" json:"encryption_type"`
	StoragePath    string             `bson:"storage_path" json:"storage_path"`
	AccessMode     string             `bson:"access_mode" json:"access_mode"`
	Whitelist      []string           `bson:"whitelist" json:"whitelist"`
	PublicCID      string             `bson:"public_cid,omitempty" json:"public_cid,omitempty"`
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time          `bson:"updated_at" json:"updated_at"`
}

func (FileMetadata) CollectionName() string {
	return "files"
}

type Notification struct {
	ID                    primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	RecipientUserID       string             `bson:"recipient_user_id" json:"recipient_user_id"`
	ActorUserID           string             `bson:"actor_user_id" json:"actor_user_id"`
	Type                  string             `bson:"type" json:"type"`
	FileID                string             `bson:"file_id" json:"file_id"`
	FileFilenameSnapshot  string             `bson:"file_filename_snapshot" json:"file_filename_snapshot"`
	ActorNameSnapshot     string             `bson:"actor_name_snapshot" json:"actor_name_snapshot"`
	ActorEmailSnapshot    string             `bson:"actor_email_snapshot" json:"actor_email_snapshot"`
	ActorAvatarSnapshot   string             `bson:"actor_avatar_snapshot" json:"actor_avatar_snapshot"`
	ReadAt                *time.Time         `bson:"read_at,omitempty" json:"read_at,omitempty"`
	CreatedAt             time.Time          `bson:"created_at" json:"created_at"`
}

func (Notification) CollectionName() string {
	return "notifications"
}

type JobType string

type JobStatus string

const (
	JobTypeEncrypt JobType = "encrypt"
	JobTypeDecrypt JobType = "decrypt"
)

const (
	JobStatusPending    JobStatus = "pending"
	JobStatusProcessing JobStatus = "processing"
	JobStatusCompleted  JobStatus = "completed"
	JobStatusFailed     JobStatus = "failed"
	JobStatusCancelled  JobStatus = "cancelled"
)

type JobError struct {
	Code      string    `bson:"code" json:"code"`
	Message   string    `bson:"message" json:"message"`
	Timestamp time.Time `bson:"timestamp" json:"timestamp"`
}

type JobResult struct {
	FileID       string `bson:"file_id,omitempty" json:"file_id,omitempty"`
	Filename     string `bson:"filename,omitempty" json:"filename,omitempty"`
	DownloadURL  string `bson:"download_url,omitempty" json:"download_url,omitempty"`
	PreviewType  string `bson:"preview_type,omitempty" json:"preview_type,omitempty"`
	MimeType     string `bson:"mime_type,omitempty" json:"mime_type,omitempty"`
	ContentBase64 string `bson:"content_base64,omitempty" json:"content_base64,omitempty"`
}

type Job struct {
	ID             primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	UserID         string              `bson:"user_id" json:"user_id"`
	Type           JobType             `bson:"type" json:"type"`
	Status         JobStatus           `bson:"status" json:"status"`
	Progress       int                 `bson:"progress" json:"progress"`
	Filename       string              `bson:"filename,omitempty" json:"filename,omitempty"`
	FileMetadataID *primitive.ObjectID `bson:"file_metadata_id,omitempty" json:"file_metadata_id,omitempty"`
	Result         *JobResult          `bson:"result,omitempty" json:"result,omitempty"`
	Error          *JobError           `bson:"error,omitempty" json:"error,omitempty"`
	CreatedAt      time.Time           `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time           `bson:"updated_at" json:"updated_at"`
	CompletedAt    *time.Time          `bson:"completed_at,omitempty" json:"completed_at,omitempty"`
}

func (Job) CollectionName() string {
	return "jobs"
}

type EncryptionType string

const (
	EncryptionAES128 EncryptionType = "AES-128"
	EncryptionAES192 EncryptionType = "AES-192"
	EncryptionAES256 EncryptionType = "AES-256"
)

type AccessMode string

const (
	AccessModePublic    AccessMode = "public"
	AccessModePrivate   AccessMode = "private"
	AccessModeWhitelist AccessMode = "whitelist"
)

const NotificationTypeFileAddedToWhitelist = "file_added_to_whitelist"

const (
	BytesPerGB            int64 = 1024 * 1024 * 1024
	MaxFileSizeBytes      int64 = 1 * BytesPerGB
	DefaultUserQuotaBytes int64 = 10 * BytesPerGB
)

type UserStorage struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID     string             `bson:"user_id" json:"user_id"`
	UsedBytes  int64              `bson:"used_bytes" json:"used_bytes"`
	FileCount  int64              `bson:"file_count" json:"file_count"`
	QuotaBytes int64              `bson:"quota_bytes" json:"quota_bytes"`
	UpdatedAt  time.Time          `bson:"updated_at" json:"updated_at"`
}

func (UserStorage) CollectionName() string {
	return "user_storage"
}
