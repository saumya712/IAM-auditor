package main

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"iam-advisor/backend/internal/auth"
	"iam-advisor/backend/internal/db"
	"iam-advisor/backend/internal/history"
	"iam-advisor/backend/internal/iam"
	"iam-advisor/backend/internal/middleware"
)

func main() {
	// Load .env file if present; ignore error when file is absent.
	_ = godotenv.Load()

	// Initialise the database connection and run migrations.
	db.Init()

	// Set up AI client.
	aiServiceURL := os.Getenv("AI_SERVICE_URL")
	if aiServiceURL == "" {
		aiServiceURL = "http://localhost:8000"
	}
	aiClient := iam.NewHTTPAIClient(aiServiceURL)
	iamHandler := iam.NewHandler(aiClient)

	router := gin.Default()

	// Apply CORS middleware globally.
	router.Use(middleware.CORSMiddleware())

	// Public auth routes.
	api := router.Group("/api")
	{
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/register", auth.Register)
			authGroup.POST("/login", auth.Login)
		}

		// Protected routes — require a valid JWT.
		protected := api.Group("/")
		protected.Use(middleware.JWTMiddleware())
		{
			protected.GET("/history", history.GetHistory)
			protected.POST("/iam/advise", iamHandler.Advise)
			protected.POST("/iam/audit", iamHandler.Audit)
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if err := router.Run(":" + port); err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}
