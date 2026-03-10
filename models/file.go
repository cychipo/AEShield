package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FileMetadata struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	OwnerID        string             `bson:"owner_id" json:"owner_id"`
	Filename       string             `bson:"filename" json:"filename"`
	Size           int64              `bson:"size" json:"size"`
	EncryptionType string             `bson:"encryption_type" json:"encryption_type"` // AES-128, AES-192, AES-256
	StoragePath    string             `bson:"storage_path" json:"storage_path"`       // Path on Cloudflare R2
	AccessMode     string             `bson:"access_mode" json:"access_mode"`         // public, private, whitelist
	Whitelist      []string           `bson:"whitelist" json:"whitelist,omitempty"`   // Allowed emails/IDs
	PublicCID      string             `bson:"public_cid,omitempty" json:"public_cid,omitempty"`
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time          `bson:"updated_at" json:"updated_at"`
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
