package main

import (
	"auth-service/internal/database"
	"auth-service/internal/handler"
	"auth-service/internal/middleware"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// 1. Load Env
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  Warning: .env file not found")
	}

	// 2. Connect DB
	database.InitDB()
	database.InitRedis()

	// 3. Setup Router
	r := gin.Default()
	r.Use(middleware.CORS())

	// 4. Routes
	auth := r.Group("/auth")
	{
		auth.POST("/register", handler.Register)
		auth.POST("/verify", handler.Verify)
		auth.POST("/login", handler.Login)
		auth.POST("/refresh", handler.Refresh)
		auth.POST("/logout", handler.Logout)
		auth.POST("/internal/send-receipt", handler.SendReceipt)

	}

	log.Println("🚀 Auth Service running on http://localhost:8080")
	r.Run(":8080")
}