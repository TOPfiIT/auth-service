package services

import (
	"context"
	"fmt"
	"time"

	"github.com/TOPfiIT/auth-service/internal/db"
	"github.com/TOPfiIT/auth-service/internal/models"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	pg             *db.PostgresDB
	redis          *db.Redis
	sessionService *SessionService
}

func NewAuthService(pg *db.PostgresDB, redis *db.Redis, ss *SessionService) *AuthService {
	return &AuthService{
		pg:             pg,
		redis:          redis,
		sessionService: ss,
	}
}

func (s *AuthService) Register(ctx context.Context, name, password string) (*models.Tokens, error) {
	exists, _ := s.pg.GetCompanyByName(ctx, name)
	if exists != nil {
		return nil, fmt.Errorf("[AuthService] company already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("[AuthService] hash password: %w", err)
	}

	company, err := s.pg.CreateCompany(ctx, name, string(hashedPassword))
	if err != nil {
		return nil, fmt.Errorf("[AuthService] create company: %w", err)
	}

	tokens, err := s.sessionService.GenerateTokens(company.ID, company.Name)
	if err != nil {
		return nil, fmt.Errorf("[AuthService] generate tokens: %w", err)
	}

	session := &models.Session{
		CompanyID:   company.ID,
		CompanyName: company.Name,
		CreatedAt:   time.Now(),
		ExpiresAt:   tokens.RefreshTokenExpireAt,
	}

	if err := s.redis.SaveSession(ctx, company.ID, session); err != nil {
		return nil, fmt.Errorf("[AuthService] save session: %w", err)
	}

	if err := s.redis.SaveRefreshToken(ctx, company.ID, tokens.RefreshToken, s.sessionService.refreshTTL); err != nil {
		return nil, fmt.Errorf("[AuthService] save refresh token: %w", err)
	}

	return tokens, nil
}

func (s *AuthService) GetCompany(ctx context.Context, accessToken string) (*models.Session, error) {
	claims, err := s.sessionService.ValidateAccessToken(accessToken)
	if err != nil {
		return nil, fmt.Errorf("[AuthService] invalid access token: %w", err)
	}

	companyID, err := uuid.Parse(claims.CompanyID)
	if err != nil {
		return nil, fmt.Errorf("[AuthService] invalid company id: %w", err)
	}

	session, err := s.redis.GetSession(ctx, companyID)
	if err != nil {
		return nil, fmt.Errorf("[AuthService] session not found: %w", err)
	}

	return session, nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*models.Tokens, error) {
	claims, err := s.sessionService.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("[AuthService] invalid refresh token: %w", err)
	}

	companyID, err := uuid.Parse(claims.CompanyID)
	if err != nil {
		return nil, fmt.Errorf("[AuthService] invalid company id: %w", err)
	}

	storedRefresh, err := s.redis.GetRefreshToken(ctx, companyID)
	if err != nil {
		return nil, fmt.Errorf("[AuthService] refresh token not found")
	}

	if storedRefresh != refreshToken {
		return nil, fmt.Errorf("[AuthService] refresh token does not equal")
	}

	session, err := s.redis.GetSession(ctx, companyID)
	if err != nil {
		return nil, fmt.Errorf("[AuthService] session not found: %w", err)
	}

	newTokens, err := s.sessionService.GenerateTokens(session.CompanyID, session.CompanyName)
	if err != nil {
		return nil, fmt.Errorf("[AuthService] generate tokens: %w", err)
	}

	session.ExpiresAt = newTokens.RefreshTokenExpireAt
	if err := s.redis.SaveSession(ctx, session.CompanyID, session); err != nil {
		return nil, fmt.Errorf("[AuthService] save session: %w", err)
	}

	if err := s.redis.SaveRefreshToken(ctx, companyID, newTokens.RefreshToken, s.sessionService.refreshTTL); err != nil {
		return nil, fmt.Errorf("[AuthService] save refresh token: %w", err)
	}

	return newTokens, nil
}

func (s *AuthService) Login(ctx context.Context, name, password string) (*models.Tokens, error) {
	company, err := s.pg.GetCompanyByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("[AuthService] company not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(company.Password), []byte(password)); err != nil {
		return nil, fmt.Errorf("[AuthService] invalid password")
	}

	tokens, err := s.sessionService.GenerateTokens(company.ID, company.Name)
	if err != nil {
		return nil, fmt.Errorf("[AuthService] generate tokens: %w", err)
	}

	session := &models.Session{
		CompanyID:   company.ID,
		CompanyName: company.Name,
		CreatedAt:   time.Now(),
		ExpiresAt:   tokens.RefreshTokenExpireAt,
	}

	if err := s.redis.SaveSession(ctx, company.ID, session); err != nil {
		return nil, fmt.Errorf("[AuthService] save session: %w", err)
	}

	if err := s.redis.SaveRefreshToken(ctx, company.ID, tokens.RefreshToken, s.sessionService.refreshTTL); err != nil {
		return nil, fmt.Errorf("[AuthService] save refresh token: %w", err)
	}

	return tokens, nil
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	claims, err := s.sessionService.ValidateRefreshToken(refreshToken)
	if err != nil {
		return fmt.Errorf("[AuthService] invalid refresh token: %w", err)
	}

	companyID, err := uuid.Parse(claims.CompanyID)
	if err != nil {
		return fmt.Errorf("[AuthService] invalid company id: %w", err)
	}

	if err := s.redis.DeleteSession(ctx, companyID); err != nil {
		return fmt.Errorf("[AuthService] delete session: %w", err)
	}

	if err := s.redis.DeleteRefreshToken(ctx, companyID); err != nil {
		return fmt.Errorf("[AuthService] delete refresh token: %w", err)
	}

	return nil
}
