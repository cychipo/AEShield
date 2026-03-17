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
