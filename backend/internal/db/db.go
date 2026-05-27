package db

import (
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"iam-advisor/backend/internal/models"
)

// DB is the shared GORM database connection.
var DB *gorm.DB

// Init initializes the PostgreSQL connection using environment variables and
// runs AutoMigrate for all models. It panics if the connection or migration fails.
func Init() {
	host := getEnv("DB_HOST", "localhost")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "iam_advisor")
	port := getEnv("DB_PORT", "5432")

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		host, user, password, dbName, port,
	)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("failed to connect to database: %v", err))
	}

	if err := DB.AutoMigrate(&models.User{}, &models.PolicyHistory{}); err != nil {
		panic(fmt.Sprintf("failed to run database migrations: %v", err))
	}
}

// getEnv returns the value of the environment variable named by key, or
// fallback if the variable is not set or is empty.
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
