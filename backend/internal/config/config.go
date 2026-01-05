package config

import (
	"os"
	"fmt"
)

type Config struct {
	// Database
	DBHost     string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string
	DatabaseURL string

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

	// BankInfo
	BankInfoEncryptionKey string
}

func Load() *Config {
	return &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBName:     getEnv("DB_NAME", "hrmsdb"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DatabaseURL:getEnv("DATABASE_URL", getDatabaseURL()),

		Port:        getEnv("PORT", "8080"),
		FrontendURL: getEnv("FRONTEND_URL", "http://localhost:5173"),

		JWTSecret: getEnv("JWT_SECRET", "change-this-secret-key"),

		AWSRegion: getEnv("AWS_REGION", "us-east-1"),
		S3Bucket:  getEnv("S3_BUCKET", "hr-app-documents"),

		Environment: getEnv("ENVIRONMENT", "development"),
		BankInfoEncryptionKey: getEnv("BANK_INFO_ENCRYPTION_KEY", "default-32-byte-long-encryption!"),
	}
}

func getDatabaseURL() string {

    dbPort := getEnv("DB_PORT", "5432") 
    dbHost := getEnv("DB_HOST", "localhost")
    dbUser := getEnv("DB_USER", "postgres")
    dbPassword := getEnv("DB_PASSWORD", "")
    dbName := getEnv("DB_NAME", "hrmsdb")

    ssl_mode := "disable" 
    if dbHost != "localhost" { 
        ssl_mode = "require" 
    }
        
    return fmt.Sprintf(
        "postgres://%s:%s@%s:%s/%s?sslmode=%s",
        dbUser,
        dbPassword,
        dbHost,
        dbPort,  // ✅ Now uses correct DB_PORT
        dbName,
        ssl_mode,
    )
	
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
