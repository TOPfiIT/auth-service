package handlers

import (
	"net/http"
	"time"

	"github.com/TOPfiIT/auth-service/internal/services"
	"github.com/TOPfiIT/auth-service/pkg/middlewares"
	"github.com/google/uuid"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	auth *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{auth: authService}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tokens, err := h.auth.Register(c.Request.Context(), req.Name, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	setTokenCookies(c, tokens.AccessToken, tokens.RefreshToken, tokens.AccessTokenExpireAt, tokens.RefreshTokenExpireAt)

	c.JSON(http.StatusCreated, gin.H{"message": "registered"})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no refresh token"})
		return
	}

	tokens, err := h.auth.Refresh(c.Request.Context(), refreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	setTokenCookies(c, tokens.AccessToken, tokens.RefreshToken, tokens.AccessTokenExpireAt, tokens.RefreshTokenExpireAt)

	c.JSON(http.StatusOK, gin.H{"message": "tokens refreshed"})
}

func (h *AuthHandler) GetCompany(c *gin.Context) {
	companyID := middlewares.GetCompanyID(c)
	companyName := middlewares.GetCompanyName(c)

	c.JSON(http.StatusOK, gin.H{
		"company_id":   companyID,
		"company_name": companyName,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tokens, err := h.auth.Login(c.Request.Context(), req.Name, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	setTokenCookies(c, tokens.AccessToken, tokens.RefreshToken, tokens.AccessTokenExpireAt, tokens.RefreshTokenExpireAt)

	c.JSON(http.StatusOK, gin.H{"message": "login succeed"})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	if refreshToken, err := c.Cookie("refresh_token"); err == nil {
		h.auth.Logout(c.Request.Context(), refreshToken)
	}

	clearTokenCookies(c)
	c.JSON(http.StatusOK, gin.H{"message": "successfully logged out"})
}

func (h *AuthHandler) CreateRoomSession(c *gin.Context) {
	var req CreateRoomSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	roomID, err := uuid.Parse(req.RoomID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	roomToken, err := h.auth.CreateRoomSession(c.Request.Context(), roomID, req.ExpiryAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, roomToken)
}

func setTokenCookies(c *gin.Context, accessToken, refreshToken string, accessExpiry, refreshExpiry time.Time) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Expires:  accessExpiry,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Expires:  refreshExpiry,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})
}

func clearTokenCookies(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		Path:     "/",
		HttpOnly: true,
	})

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		Path:     "/",
		HttpOnly: true,
	})
}
