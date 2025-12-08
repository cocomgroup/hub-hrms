package config

import (
	"os"
)

type Config struct {
	// Database
	DBHost     string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string

	// Server
	Port        string
	FrontendURL string

	// JWT
	JWTSecret string

	// AWS
	AWSRegion string
	S3Bucket  string

	// Application
	Environment string
}

func Load() *Config {
	return &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBName:     getEnv("DB_NAME", "hrmsdb"),
		DBUser:     getEnv("DB_USER", "hrms_user"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),

		Port:        getEnv("PORT", "8080"),
		FrontendURL: getEnv("FRONTEND_URL", "http://localhost:5173"),

		JWTSecret: getEnv("JWT_SECRET", "change-this-secret-key"),

		AWSRegion: getEnv("AWS_REGION", "us-east-1"),
		S3Bucket:  getEnv("S3_BUCKET", "hr-app-documents"),

		Environment: getEnv("ENVIRONMENT", "development"),
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
