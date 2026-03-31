// Package models định nghĩa các cấu trúc dữ liệu liên quan đến quản lý quyền truy cập
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AccessRule đại diện cho một quy tắc truy cập
type AccessRule struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ResourceID  string             `bson:"resource_id" json:"resource_id"`       // ID của tài nguyên cần kiểm soát
	ResourceType string            `bson:"resource_type" json:"resource_type"`   // Loại tài nguyên (file, folder, etc.)
	OwnerID     string             `bson:"owner_id" json:"owner_id"`            // ID của chủ sở hữu
	AccessMode  AccessMode         `bson:"access_mode" json:"access_mode"`       // Chế độ truy cập
	Whitelist   []string           `bson:"whitelist" json:"whitelist,omitempty"` // Danh sách người dùng được phép
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

// AccessMode là kiểu dữ liệu cho chế độ truy cập
type AccessMode string

const (
	// AccessModePublic cho phép bất kỳ ai có link truy cập
	AccessModePublic AccessMode = "public"

	// AccessModePrivate chỉ chủ sở hữu được truy cập
	AccessModePrivate AccessMode = "private"

	// AccessModeWhitelist chỉ những người trong danh sách trắng được truy cập
	AccessModeWhitelist AccessMode = "whitelist"
)

// CheckAccessRequest chứa thông tin để kiểm tra quyền truy cập
type CheckAccessRequest struct {
	ResourceID    string // ID của tài nguyên
	RequesterID   string // ID của người yêu cầu (nếu có xác thực)
	RequesterEmail string // Email của người yêu cầu (nếu có)
}

// CheckAccessResult kết quả kiểm tra quyền truy cập
type CheckAccessResult struct {
	HasAccess bool   // Có quyền truy cập hay không
	Reason    string // Lý do (cho mục đích ghi log)
	Error     error  // Lỗi nếu có
}