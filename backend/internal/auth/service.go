package auth

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"strconv"
	"time"

	"github.com/aeshield/backend/internal/config"
	"github.com/aeshield/backend/internal/database"
	"github.com/aeshield/backend/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

type UserRepository interface {
	FindByProvider(ctx context.Context, provider, providerID string) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	Create(ctx context.Context, user *models.User) error
	Update(ctx context.Context, user *models.User) error
}

type GoogleUser struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

type GitHubUser struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

type Service struct {
	cfg       *config.Config
	userRepo  UserRepository
	googleCfg *oauth2.Config
	githubCfg *oauth2.Config
}

func NewService(cfg *config.Config, userRepo UserRepository) *Service {
	return &Service{
		cfg:      cfg,
		userRepo: userRepo,
		googleCfg: &oauth2.Config{
			ClientID:     cfg.GoogleClientID,
			ClientSecret: cfg.GoogleClientSecret,
			Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
			Endpoint:     google.Endpoint,
			RedirectURL:  cfg.GoogleRedirectURL,
		},
		githubCfg: &oauth2.Config{
			ClientID:     cfg.GitHubClientID,
			ClientSecret: cfg.GitHubClientSecret,
			Scopes:       []string{"user:email", "read:user"},
			Endpoint:     github.Endpoint,
			RedirectURL:  cfg.GitHubRedirectURL,
		},
	}
}

func (s *Service) GoogleAuthURL() string {
	return s.googleCfg.AuthCodeURL(uuid.New().String())
}

func (s *Service) GitHubAuthURL() string {
	return s.githubCfg.AuthCodeURL(uuid.New().String())
}

func (s *Service) ExchangeGoogleCode(code string) (*oauth2.Token, error) {
	token, err := s.googleCfg.Exchange(context.Background(), code)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (s *Service) ExchangeGitHubCode(code string) (*oauth2.Token, error) {
	token, err := s.githubCfg.Exchange(context.Background(), code)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (s *Service) GetGoogleUserInfo(token *oauth2.Token) (*models.Claims, error) {
	client := s.googleCfg.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return s.parseGoogleUserInfo(resp.Body)
}

func (s *Service) GetGitHubUserInfo(token *oauth2.Token) (*models.Claims, error) {
	client := s.githubCfg.Client(context.Background(), token)
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	claims, err := s.parseGitHubUserInfo(resp.Body)
	if err != nil {
		return nil, err
	}

	if claims.Email == "" {
		email, err := s.getGitHubPrimaryEmail(token)
		if err == nil && email != "" {
			claims.Email = email
		}
	}

	return claims, nil
}

func (s *Service) getGitHubPrimaryEmail(token *oauth2.Token) (string, error) {
	client := s.githubCfg.Client(context.Background(), token)
	resp, err := client.Get("https://api.github.com/user/emails")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	type GitHubEmail struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}

	var emails []GitHubEmail
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", err
	}

	for _, e := range emails {
		if e.Primary && e.Verified {
			return e.Email, nil
		}
	}

	return "", errors.New("no primary email found")
}

func (s *Service) FindOrCreateUser(ctx context.Context, provider, providerID, email, name, avatar string) (*models.User, error) {
	if s.userRepo == nil {
		return nil, errors.New("user repository not configured")
	}

	// 1. Tìm theo provider + providerID (đã từng login trước đó)
	user, err := s.userRepo.FindByProvider(ctx, provider, providerID)
	if err == nil {
		// Cập nhật thông tin mới nhất
		updated := false
		if user.Name != name {
			user.Name = name
			updated = true
		}
		if user.Avatar != avatar {
			user.Avatar = avatar
			updated = true
		}
		if updated {
			if err := s.userRepo.Update(ctx, user); err != nil {
				return nil, err
			}
		}
		return user, nil
	}

	if !errors.Is(err, database.ErrUserNotFound) {
		return nil, err
	}

	// 2. Tìm theo email - nếu cùng email thì merge provider vào user hiện có
	if email != "" {
		existingUser, err := s.userRepo.FindByEmail(ctx, email)
		if err == nil {
			// Thêm provider mới vào user đã có
			existingUser.Providers = append(existingUser.Providers, models.LinkedProvider{
				Provider:   provider,
				ProviderID: providerID,
			})
			// Cập nhật avatar nếu chưa có
			if existingUser.Avatar == "" && avatar != "" {
				existingUser.Avatar = avatar
			}
			if err := s.userRepo.Update(ctx, existingUser); err != nil {
				return nil, err
			}
			return existingUser, nil
		}
		if !errors.Is(err, database.ErrUserNotFound) {
			return nil, err
		}
	}

	// 3. Tạo user mới
	user = &models.User{
		Email:  email,
		Name:   name,
		Avatar: avatar,
		Providers: []models.LinkedProvider{
			{Provider: provider, ProviderID: providerID},
		},
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Service) GenerateJWT(claims *models.Claims) (string, error) {
	claims.ExpiresAt = time.Now().Add(24 * time.Hour).Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWTSecret))
}

func (s *Service) parseGoogleUserInfo(body io.Reader) (*models.Claims, error) {
	var user GoogleUser
	if err := json.NewDecoder(body).Decode(&user); err != nil {
		return nil, err
	}

	return &models.Claims{
		UserID:   user.ID,
		Email:    user.Email,
		Name:     user.Name,
		Avatar:   user.Picture,
		Provider: "google",
	}, nil
}

func (s *Service) parseGitHubUserInfo(body io.Reader) (*models.Claims, error) {
	var user GitHubUser
	if err := json.NewDecoder(body).Decode(&user); err != nil {
		return nil, err
	}

	return &models.Claims{
		UserID:   "github:" + strconv.FormatInt(user.ID, 10),
		Email:    user.Email,
		Name:     user.Name,
		Avatar:   user.AvatarURL,
		Provider: "github",
	}, nil
}
