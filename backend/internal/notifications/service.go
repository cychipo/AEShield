package notifications

import (
	"context"
	"time"

	"github.com/aeshield/backend/internal/storage"
)

type Repository interface {
	ListByRecipient(ctx context.Context, recipientUserID string, limit int64, cursor string) (*storage.NotificationListResult, error)
	CountUnreadByRecipient(ctx context.Context, recipientUserID string) (int64, error)
	MarkAllRead(ctx context.Context, recipientUserID string, readAt time.Time) (int64, error)
}

type Service struct {
	repo Repository
}

type ListResponse struct {
	Items       []NotificationItem `json:"items"`
	HasMore     bool               `json:"has_more"`
	NextCursor  string             `json:"next_cursor,omitempty"`
	UnreadCount int64              `json:"unread_count"`
}

type NotificationItem struct {
	ID        string            `json:"id"`
	Type      string            `json:"type"`
	File      NotificationFile  `json:"file"`
	Actor     NotificationActor `json:"actor"`
	IsRead    bool              `json:"is_read"`
	ReadAt    *time.Time        `json:"read_at,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
}

type NotificationFile struct {
	ID       string `json:"id"`
	Filename string `json:"filename"`
}

type NotificationActor struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Avatar string `json:"avatar"`
}

type MarkAllReadResponse struct {
	UpdatedCount int64     `json:"updated_count"`
	ReadAt       time.Time `json:"read_at"`
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, recipientUserID string, limit int64, cursor string) (*ListResponse, error) {
	result, err := s.repo.ListByRecipient(ctx, recipientUserID, limit, cursor)
	if err != nil {
		return nil, err
	}

	unreadCount, err := s.repo.CountUnreadByRecipient(ctx, recipientUserID)
	if err != nil {
		return nil, err
	}

	items := make([]NotificationItem, 0, len(result.Items))
	for _, item := range result.Items {
		if item == nil {
			continue
		}
		items = append(items, NotificationItem{
			ID:   item.ID.Hex(),
			Type: item.Type,
			File: NotificationFile{
				ID:       item.FileID,
				Filename: item.FileFilenameSnapshot,
			},
			Actor: NotificationActor{
				ID:     item.ActorUserID,
				Name:   item.ActorNameSnapshot,
				Email:  item.ActorEmailSnapshot,
				Avatar: item.ActorAvatarSnapshot,
			},
			IsRead:    item.ReadAt != nil,
			ReadAt:    item.ReadAt,
			CreatedAt: item.CreatedAt,
		})
	}

	return &ListResponse{
		Items:       items,
		HasMore:     result.HasMore,
		NextCursor:  result.NextCursor,
		UnreadCount: unreadCount,
	}, nil
}

func (s *Service) MarkAllRead(ctx context.Context, recipientUserID string) (*MarkAllReadResponse, error) {
	readAt := time.Now().UTC()
	updatedCount, err := s.repo.MarkAllRead(ctx, recipientUserID, readAt)
	if err != nil {
		return nil, err
	}

	return &MarkAllReadResponse{
		UpdatedCount: updatedCount,
		ReadAt:       readAt,
	}, nil
}
