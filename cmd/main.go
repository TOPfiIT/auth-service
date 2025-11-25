package main

import (
	"fmt"
	"log"

	"github.com/TOPfiIT/auth-service/internal/config"
	"github.com/TOPfiIT/auth-service/internal/db"
	"github.com/TOPfiIT/auth-service/internal/http/handlers"
	"github.com/TOPfiIT/auth-service/internal/services"

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

	// Setup router
	r := gin.Default()

	// Register routes
	r.POST("/register", authHandler.Register)
	r.POST("/login", authHandler.Login)
	r.POST("/refresh", authHandler.Refresh)
	r.POST("/company", authHandler.GetCompany)
	r.POST("/logout", authHandler.Logout)
	r.POST("/room/session", authHandler.CreateRoomSession)

	log.Println("✓ Routes registered")

	// Start server
	addr := fmt.Sprintf(":%d", cfg.Port)
	fmt.Printf("Server starting on %s", addr)

	if err := r.Run(addr); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
