package main

import (
	"log"

	"github.com/TOPfiIT/auth-service/internal/config"
	"github.com/TOPfiIT/auth-service/internal/db"
	"github.com/TOPfiIT/auth-service/internal/http/handlers"
	"github.com/TOPfiIT/auth-service/internal/services"

	"github.com/gin-gonic/gin"
)

func main() {
	//init config
	cfg := config.MustLoad()

	//init postgres
	pg := db.InitPostgres(cfg)
	defer pg.Close()

	//init redis
	redis := db.InitRedisClient(cfg)
	defer redis.Close()

	//init services
	SessionService, err := services.NewSessionService(
		cfg.JWT.AccessTTLMinutes,
		cfg.JWT.RefreshTTLHours,
	)
	if err != nil {
		log.Fatalf("Failed to create session service: %v", err)
	}

	AuthService := services.NewAuthService(pg, redis, SessionService)

	//init handlers
	AuthHandler := handlers.NewAuthHandler(AuthService)

	//create router
	r := gin.Default()

	//create routes
	r.POST("/register", AuthHandler.Register)
	r.POST("/refresh", AuthHandler.Refresh)
	r.POST("/login", AuthHandler.Login)
	r.POST("/logout", AuthHandler.Logout)
	r.GET("/company", AuthHandler.GetCompany)

	//start server
	if err := r.Run(":8087"); err != nil {
		log.Fatal(err)
	}
}
