package services

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"time"

	"github.com/TOPfiIT/auth-service/internal/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrEmptyToken = fmt.Errorf("empty token")
)

type SessionService struct {
	privateKey *ecdsa.PrivateKey
	publicKey  *ecdsa.PublicKey
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewSessionService(accessTTLMinutes, refreshTTLHours int) (*SessionService, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("[SessionService] generate key pair: %w", err)
	}

	return &SessionService{
		privateKey: privateKey,
		publicKey:  &privateKey.PublicKey,
		accessTTL:  time.Duration(accessTTLMinutes) * time.Minute,
		refreshTTL: time.Duration(refreshTTLHours) * time.Hour,
	}, nil
}

func (s *SessionService) GenerateTokens(companyID uuid.UUID, companyName string) (*models.Tokens, error) {
	const op = "[SessionService.GenerateTokens]"

	accessToken := jwt.New(jwt.SigningMethodES256)
	claims := accessToken.Claims.(jwt.MapClaims)
	claims["cid"] = companyID.String()
	claims["cname"] = companyName
	claims["type"] = "access"
	claims["exp"] = time.Now().Add(s.accessTTL).Unix()
	claims["iat"] = time.Now().Unix()

	signedAccessToken, err := accessToken.SignedString(s.privateKey)
	if err != nil {
		return nil, fmt.Errorf("%s %w", op, err)
	}

	refreshToken := jwt.New(jwt.SigningMethodES256)
	refreshClaims := refreshToken.Claims.(jwt.MapClaims)
	refreshClaims["cid"] = companyID.String()
	refreshClaims["cname"] = companyName
	refreshClaims["type"] = "refresh"
	refreshClaims["exp"] = time.Now().Add(s.refreshTTL).Unix()
	refreshClaims["iat"] = time.Now().Unix()

	signedRefreshToken, err := refreshToken.SignedString(s.privateKey)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &models.Tokens{
		AccessToken:          signedAccessToken,
		RefreshToken:         signedRefreshToken,
		AccessTokenExpireAt:  time.Now().Add(s.accessTTL),
		RefreshTokenExpireAt: time.Now().Add(s.refreshTTL),
	}, nil
}

func (s *SessionService) NewTokens(accessToken, refreshToken string) (*models.Tokens, error) {
	if accessToken == "" || refreshToken == "" {
		return nil, ErrEmptyToken
	}

	accessClaims, err := s.parseToken(accessToken)
	if err != nil {
		return nil, err
	}

	refreshClaims, err := s.parseToken(refreshToken)
	if err != nil {
		return nil, err
	}

	accessExp := time.Unix(int64(accessClaims["exp"].(float64)), 0)
	refreshExp := time.Unix(int64(refreshClaims["exp"].(float64)), 0)

	return &models.Tokens{
		AccessToken:          accessToken,
		RefreshToken:         refreshToken,
		AccessTokenExpireAt:  accessExp,
		RefreshTokenExpireAt: refreshExp,
	}, nil
}

func (s *SessionService) ValidateAccessToken(tokenString string) (*models.TokenClaims, error) {
	const op = "[SessionService.ValidateAccessToken]"

	claims, err := s.parseToken(tokenString)
	if err != nil {
		return nil, fmt.Errorf("%s %w", op, err)
	}

	tokenType, ok := claims["type"].(string)
	if !ok || tokenType != "access" {
		return nil, fmt.Errorf("%s invalid token type", op)
	}

	companyID, ok := claims["cid"].(string)
	if !ok {
		return nil, fmt.Errorf("%s missing company id", op)
	}

	companyName, ok := claims["cname"].(string)
	if !ok {
		return nil, fmt.Errorf("%s missing company name", op)
	}

	return &models.TokenClaims{
		CompanyID:   companyID,
		CompanyName: companyName,
		TokenType:   tokenType,
	}, nil
}

func (s *SessionService) ValidateRefreshToken(tokenString string) (*models.TokenClaims, error) {
	const op = "[SessionService.ValidateRefreshToken]"

	claims, err := s.parseToken(tokenString)
	if err != nil {
		return nil, fmt.Errorf("%s %w", op, err)
	}

	tokenType, ok := claims["type"].(string)
	if !ok || tokenType != "refresh" {
		return nil, fmt.Errorf("%s invalid token type", op)
	}

	companyID, ok := claims["cid"].(string)
	if !ok {
		return nil, fmt.Errorf("%s missing company id", op)
	}

	companyName, ok := claims["cname"].(string)
	if !ok {
		return nil, fmt.Errorf("%s missing company name", op)
	}

	return &models.TokenClaims{
		CompanyID:   companyID,
		CompanyName: companyName,
		TokenType:   tokenType,
	}, nil
}

func (s *SessionService) GetPublicKey() *ecdsa.PublicKey {
	return s.publicKey
}

func (s *SessionService) parseToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.publicKey, nil
	})
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims format")
	}

	return claims, nil
}

func (s *SessionService) GenerateSessionToken(roomID uuid.UUID, expiryAt time.Time) (string, error) {
	const op = "[SessionService.GenerateSessionToken]"

	token := jwt.New(jwt.SigningMethodES256)
	claims := token.Claims.(jwt.MapClaims)
	claims["room_id"] = roomID.String()
	claims["expiry_at"] = expiryAt.Unix()
	claims["exp"] = expiryAt.Unix()
	claims["iat"] = time.Now().Unix()
	claims["type"] = "room"

	roomToken, err := token.SignedString(s.privateKey)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return roomToken, nil
}
