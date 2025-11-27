package middlewares

import (
	"net/http"

	"github.com/TOPfiIT/auth-service/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(sessionService *services.SessionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/register" || c.Request.URL.Path == "/login" ||
			c.Request.URL.Path == "/refresh" || c.Request.URL.Path == "/room/session" {
			c.Next()
			return
		}

		tokenString, err := c.Cookie("access_token")
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "no access token"})
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return sessionService.GetPublicKey(), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
			c.Abort()
			return
		}

		companyID, ok := claims["cid"].(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "company id not found"})
			c.Abort()
			return
		}

		companyName, _ := claims["cname"].(string)

		c.Set("company_id", companyID)
		c.Set("company_name", companyName)
		c.Next()
	}
}

func GetCompanyID(c *gin.Context) string {
	companyID, _ := c.Get("company_id")
	return companyID.(string)
}

func GetCompanyName(c *gin.Context) string {
	companyName, _ := c.Get("company_name")
	if name, ok := companyName.(string); ok {
		return name
	}
	return ""
}
