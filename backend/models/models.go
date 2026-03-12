package models

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	Provider  string `json:"provider"`
	ExpiresAt int64  `json:"exp"`
	Avatar    string `json:"avatar"`
	Name      string `json:"name"`
}

func (c *Claims) GetExpirationTime() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(time.Unix(c.ExpiresAt, 0)), nil
}

func (c *Claims) GetIssuedAt() (*jwt.NumericDate, error) {
	return nil, nil
}

func (c *Claims) GetNotBefore() (*jwt.NumericDate, error) {
	return nil, nil
}

func (c *Claims) GetIssuer() (string, error) {
	return "aeshield", nil
}

func (c *Claims) GetSubject() (string, error) {
	return c.UserID, nil
}

func (c *Claims) GetAudience() (jwt.ClaimStrings, error) {
	return nil, nil
}

type TokenRequest struct {
	Code string `json:"code"`
}
