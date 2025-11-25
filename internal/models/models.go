package models

import (
	"time"

	"github.com/google/uuid"
)

type Company struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
}

type Session struct {
	CompanyID   uuid.UUID `json:"company_id"`
	CompanyName string    `json:"company_name"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at"`
}

type Tokens struct {
	AccessToken          string    `json:"access_token"`
	RefreshToken         string    `json:"refresh_token"`
	AccessTokenExpireAt  time.Time `json:"access_token_expire_at"`
	RefreshTokenExpireAt time.Time `json:"refresh_token_expire_at"`
}

type TokenClaims struct {
	CompanyID   string `json:"cid"`
	CompanyName string `json:"cname"`
	TokenType   string `json:"type"`
}
