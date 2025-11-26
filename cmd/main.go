package main

import (
	"fmt"
	"log"

	"github.com/TOPfiIT/auth-service/internal/config"
	"github.com/TOPfiIT/auth-service/internal/db"
	"github.com/TOPfiIT/auth-service/internal/http/handlers"
	"github.com/TOPfiIT/auth-service/internal/services"
	"github.com/TOPfiIT/auth-service/pkg/middlewares"

	"github.com/gin-gonic/gin"
)

func main() {
	log.Println("Starting auth service")

	defer func() {
		if r := recover(); r != nil {
			log.Printf("PANIC recovered: %v", r)
		}
	}()

	cfg := config.MustLoad()
	log.Println("✓ Config loaded")

	pgDB := db.InitPostgres(cfg)
	log.Println("✓ PostgreSQL connected")

	redisClient := db.InitRedisClient(cfg)
	log.Println("✓ Redis connected")

	sessionSvc, err := services.NewSessionService(cfg)
	if err != nil {
		log.Fatalf("Failed to create session service: %v", err)
	}
	log.Println("✓ Session service created")

	authSvc := services.NewAuthService(pgDB, redisClient, sessionSvc)
	log.Println("✓ Auth service created")

	authHandler := handlers.NewAuthHandler(authSvc)
	log.Println("✓ Handlers created")

	authMiddleware := middlewares.AuthMiddleware(sessionSvc)

	// Setup router
	r := gin.Default()

	// r.Use(CORSMiddleware())

	// Register routes
	r.POST("/register", authHandler.Register)
	r.POST("/login", authHandler.Login)
	r.POST("/refresh", authHandler.Refresh)
	r.GET("/company", authMiddleware, authHandler.GetCompany)
	r.POST("/logout", authMiddleware, authHandler.Logout)
	r.POST("/room/session", authHandler.CreateRoomSession)

	log.Println("✓ Routes registered")

	// Start server
	addr := fmt.Sprintf(":%d", cfg.Port)
	fmt.Printf("Server starting on %s", addr)

	if err := r.Run(addr); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

// func CORSMiddleware() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
// 		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
// 		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
// 		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

// 		if c.Request.Method == "OPTIONS" {
// 			c.AbortWithStatus(204)
// 			return
// 		}

// 		c.Next()
// 	}
// }
