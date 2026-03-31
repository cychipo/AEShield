// Package repository định nghĩa các phương thức truy cập dữ liệu cho hệ thống quản lý quyền truy cập
package repository

import (
	"context"

	"github.com/aeshield/backend/internal/accesscontrol/models"
)

// AccessControlRepository là interface cho việc lưu trữ và truy xuất quy tắc truy cập
type AccessControlRepository interface {
	// CreateRule tạo mới một quy tắc truy cập
	CreateRule(ctx context.Context, rule *models.AccessRule) error

	// GetRuleByID lấy quy tắc truy cập theo ID
	GetRuleByID(ctx context.Context, id interface{}) (*models.AccessRule, error)

	// GetRuleByResource lấy quy tắc truy cập theo ID tài nguyên
	GetRuleByResource(ctx context.Context, resourceID string, resourceTypes ...string) (*models.AccessRule, error)

	// UpdateRule cập nhật một quy tắc truy cập
	UpdateRule(ctx context.Context, id interface{}, rule *models.AccessRule) error

	// UpdateRuleByResource cập nhật quy tắc truy cập theo ID tài nguyên
	UpdateRuleByResource(ctx context.Context, resourceID string, rule *models.AccessRule) error

	// DeleteRule xóa một quy tắc truy cập
	DeleteRule(ctx context.Context, id interface{}) error

	// DeleteRuleByResource xóa quy tắc truy cập theo ID tài nguyên
	DeleteRuleByResource(ctx context.Context, resourceID string) error

	// IsOwner kiểm tra xem người dùng có phải là chủ sở hữu của tài nguyên hay không
	IsOwner(ctx context.Context, resourceID, userID string) (bool, error)

	// GetWhitelist lấy danh sách trắng của một tài nguyên
	GetWhitelist(ctx context.Context, resourceID string) ([]string, error)
}