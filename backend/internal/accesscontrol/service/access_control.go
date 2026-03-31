// Package service cung cấp logic nghiệp vụ cho hệ thống quản lý quyền truy cập
package service

import (
	"context"
	"fmt"
	"time"

	"github.com/aeshield/backend/internal/accesscontrol/models"
	"github.com/aeshield/backend/internal/accesscontrol/repository"
)

// AccessControlService cung cấp các phương thức để quản lý quyền truy cập
type AccessControlService struct {
	repo repository.AccessControlRepository
}

// NewAccessControlService tạo mới một instance của AccessControlService
func NewAccessControlService(repo repository.AccessControlRepository) *AccessControlService {
	return &AccessControlService{
		repo: repo,
	}
}

// CreateRule tạo mới một quy tắc truy cập
func (s *AccessControlService) CreateRule(ctx context.Context, resourceID, resourceType, ownerID string, accessMode models.AccessMode, whitelist []string) (*models.AccessRule, error) {
	rule := &models.AccessRule{
		ResourceID:   resourceID,
		ResourceType: resourceType,
		OwnerID:      ownerID,
		AccessMode:   accessMode,
		Whitelist:    whitelist,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := s.repo.CreateRule(ctx, rule)
	if err != nil {
		return nil, fmt.Errorf("failed to create access rule: %w", err)
	}

	return rule, nil
}

// GetRuleByResource lấy quy tắc truy cập theo ID tài nguyên
func (s *AccessControlService) GetRuleByResource(ctx context.Context, resourceID string, resourceTypes ...string) (*models.AccessRule, error) {
	rule, err := s.repo.GetRuleByResource(ctx, resourceID, resourceTypes...)
	if err != nil {
		return nil, fmt.Errorf("failed to get access rule: %w", err)
	}
	return rule, nil
}

// UpdateRule cập nhật quy tắc truy cập
func (s *AccessControlService) UpdateRule(ctx context.Context, resourceID string, accessMode models.AccessMode, whitelist []string) (*models.AccessRule, error) {
	rule, err := s.GetRuleByResource(ctx, resourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing rule: %w", err)
	}

	rule.AccessMode = accessMode
	rule.Whitelist = whitelist
	rule.UpdatedAt = time.Now()

	err = s.repo.UpdateRuleByResource(ctx, resourceID, rule)
	if err != nil {
		return nil, fmt.Errorf("failed to update access rule: %w", err)
	}

	return rule, nil
}

// DeleteRule xóa quy tắc truy cập
func (s *AccessControlService) DeleteRule(ctx context.Context, resourceID string) error {
	err := s.repo.DeleteRuleByResource(ctx, resourceID)
	if err != nil {
		return fmt.Errorf("failed to delete access rule: %w", err)
	}
	return nil
}

// CheckAccess kiểm tra quyền truy cập cho một tài nguyên
func (s *AccessControlService) CheckAccess(ctx context.Context, req *models.CheckAccessRequest) *models.CheckAccessResult {
	rule, err := s.GetRuleByResource(ctx, req.ResourceID)
	if err != nil {
		return &models.CheckAccessResult{
			HasAccess: false,
			Reason:    "Access rule not found",
			Error:     err,
		}
	}

	switch rule.AccessMode {
	case models.AccessModePublic:
		// Ai có link cũng có thể truy cập
		return &models.CheckAccessResult{
			HasAccess: true,
			Reason:    "Public resource",
		}

	case models.AccessModePrivate:
		// Chỉ chủ sở hữu được truy cập
		isOwner, err := s.repo.IsOwner(ctx, req.ResourceID, req.RequesterID)
		if err != nil {
			return &models.CheckAccessResult{
				HasAccess: false,
				Reason:    "Error checking ownership",
				Error:     err,
			}
		}

		if isOwner {
			return &models.CheckAccessResult{
				HasAccess: true,
				Reason:    "Owner access",
			}
		} else {
			return &models.CheckAccessResult{
				HasAccess: false,
				Reason:    "Access denied: private resource, not owner",
			}
		}

	case models.AccessModeWhitelist:
		// Kiểm tra xem người dùng có trong danh sách trắng không
		if req.RequesterID != "" {
			for _, userID := range rule.Whitelist {
				if userID == req.RequesterID {
					return &models.CheckAccessResult{
						HasAccess: true,
						Reason:    "User in whitelist",
					}
				}
			}
		}

		if req.RequesterEmail != "" {
			for _, email := range rule.Whitelist {
				if email == req.RequesterEmail {
					return &models.CheckAccessResult{
						HasAccess: true,
						Reason:    "Email in whitelist",
					}
				}
			}
		}

		return &models.CheckAccessResult{
			HasAccess: false,
			Reason:    "Access denied: not in whitelist",
		}

	default:
		return &models.CheckAccessResult{
			HasAccess: false,
			Reason:    "Unknown access mode",
		}
	}
}

// IsOwner kiểm tra xem người dùng có phải là chủ sở hữu của tài nguyên không
func (s *AccessControlService) IsOwner(ctx context.Context, resourceID, userID string) (bool, error) {
	isOwner, err := s.repo.IsOwner(ctx, resourceID, userID)
	if err != nil {
		return false, fmt.Errorf("failed to check ownership: %w", err)
	}
	return isOwner, nil
}

// GetWhitelist lấy danh sách trắng của một tài nguyên
func (s *AccessControlService) GetWhitelist(ctx context.Context, resourceID string) ([]string, error) {
	whitelist, err := s.repo.GetWhitelist(ctx, resourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get whitelist: %w", err)
	}
	return whitelist, nil
}

// AddToWhitelist thêm người dùng vào danh sách trắng
func (s *AccessControlService) AddToWhitelist(ctx context.Context, resourceID, userIdentifier string) error {
	rule, err := s.GetRuleByResource(ctx, resourceID)
	if err != nil {
		return fmt.Errorf("failed to get rule: %w", err)
	}

	// Kiểm tra xem đã tồn tại chưa
	for _, item := range rule.Whitelist {
		if item == userIdentifier {
			return fmt.Errorf("user already in whitelist")
		}
	}

	// Thêm vào danh sách trắng
	rule.Whitelist = append(rule.Whitelist, userIdentifier)
	rule.UpdatedAt = time.Now()

	err = s.repo.UpdateRuleByResource(ctx, resourceID, rule)
	if err != nil {
		return fmt.Errorf("failed to update whitelist: %w", err)
	}

	return nil
}

// RemoveFromWhitelist xóa người dùng khỏi danh sách trắng
func (s *AccessControlService) RemoveFromWhitelist(ctx context.Context, resourceID, userIdentifier string) error {
	rule, err := s.GetRuleByResource(ctx, resourceID)
	if err != nil {
		return fmt.Errorf("failed to get rule: %w", err)
	}

	// Xóa khỏi danh sách trắng
	newWhitelist := []string{}
	for _, item := range rule.Whitelist {
		if item != userIdentifier {
			newWhitelist = append(newWhitelist, item)
		}
	}

	rule.Whitelist = newWhitelist
	rule.UpdatedAt = time.Now()

	err = s.repo.UpdateRuleByResource(ctx, resourceID, rule)
	if err != nil {
		return fmt.Errorf("failed to update whitelist: %w", err)
	}

	return nil
}