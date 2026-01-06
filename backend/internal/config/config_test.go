package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		validate func(*testing.T, *Config)
	}{
		{
			name: "loads default config",
			envVars: map[string]string{},
			validate: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "localhost", cfg.DBHost)
				assert.Equal(t, "5432", cfg.DBPort)
				assert.Equal(t, "hrmsdb", cfg.DBName)
				assert.Equal(t, "postgres", cfg.DBUser)
				assert.Equal(t, "8080", cfg.Port)
				assert.Equal(t, "http://localhost:5173", cfg.FrontendURL)
				assert.Equal(t, "development", cfg.Environment)
			},
		},
		{
			name: "loads custom config from environment",
			envVars: map[string]string{
				"DB_HOST":     "db.example.com",
				"DB_PORT":     "5433",
				"DB_NAME":     "testdb",
				"DB_USER":     "testuser",
				"DB_PASSWORD": "testpass",
				"PORT":        "9000",
				"FRONTEND_URL": "https://app.example.com",
				"JWT_SECRET":  "test-secret-key",
				"AWS_REGION":  "us-west-2",
				"S3_BUCKET":   "test-bucket",
				"ENVIRONMENT": "production",
				"BANK_INFO_ENCRYPTION_KEY": "test-encryption-key-32-chars!",
			},
			validate: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "db.example.com", cfg.DBHost)
				assert.Equal(t, "5433", cfg.DBPort)
				assert.Equal(t, "testdb", cfg.DBName)
				assert.Equal(t, "testuser", cfg.DBUser)
				assert.Equal(t, "testpass", cfg.DBPassword)
				assert.Equal(t, "9000", cfg.Port)
				assert.Equal(t, "https://app.example.com", cfg.FrontendURL)
				assert.Equal(t, "test-secret-key", cfg.JWTSecret)
				assert.Equal(t, "us-west-2", cfg.AWSRegion)
				assert.Equal(t, "test-bucket", cfg.S3Bucket)
				assert.Equal(t, "production", cfg.Environment)
				assert.Equal(t, "test-encryption-key-32-chars!", cfg.BankInfoEncryptionKey)
			},
		},
		{
			name: "loads DATABASE_URL when provided",
			envVars: map[string]string{
				"DATABASE_URL": "postgres://user:pass@remotehost:5432/db?sslmode=require",
			},
			validate: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "postgres://user:pass@remotehost:5432/db?sslmode=require", cfg.DatabaseURL)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original env vars
			originalEnv := saveEnv()
			defer restoreEnv(originalEnv)

			// Set test env vars
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			// Load config
			cfg := Load()

			// Validate
			assert.NotNil(t, cfg)
			tt.validate(t, cfg)
		})
	}
}

func TestGetDatabaseURL(t *testing.T) {
	tests := []struct {
		name        string
		envVars     map[string]string
		expectedURL string
	}{
		{
			name: "localhost with default values",
			envVars: map[string]string{
				"DB_HOST": "localhost",
				"DB_PORT": "5432",
				"DB_USER": "postgres",
				"DB_PASSWORD": "testpass",
				"DB_NAME": "testdb",
			},
			expectedURL: "postgres://postgres:testpass@localhost:5432/testdb?sslmode=disable",
		},
		{
			name: "remote host with SSL",
			envVars: map[string]string{
				"DB_HOST": "db.example.com",
				"DB_PORT": "5432",
				"DB_USER": "dbuser",
				"DB_PASSWORD": "dbpass",
				"DB_NAME": "proddb",
			},
			expectedURL: "postgres://dbuser:dbpass@db.example.com:5432/proddb?sslmode=require",
		},
		{
			name: "custom port",
			envVars: map[string]string{
				"DB_HOST": "localhost",
				"DB_PORT": "5433",
				"DB_USER": "testuser",
				"DB_PASSWORD": "testpass",
				"DB_NAME": "testdb",
			},
			expectedURL: "postgres://testuser:testpass@localhost:5433/testdb?sslmode=disable",
		},
		{
			name: "empty password",
			envVars: map[string]string{
				"DB_HOST": "localhost",
				"DB_PORT": "5432",
				"DB_USER": "postgres",
				"DB_PASSWORD": "",
				"DB_NAME": "testdb",
			},
			expectedURL: "postgres://postgres:@localhost:5432/testdb?sslmode=disable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original env vars
			originalEnv := saveEnv()
			defer restoreEnv(originalEnv)

			// Set test env vars
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			// Get database URL
			url := getDatabaseURL()

			// Validate
			assert.Equal(t, tt.expectedURL, url)
		})
	}
}

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		expected     string
	}{
		{
			name:         "returns environment value when set",
			key:          "TEST_VAR",
			defaultValue: "default",
			envValue:     "custom",
			expected:     "custom",
		},
		{
			name:         "returns default value when not set",
			key:          "UNSET_VAR",
			defaultValue: "default",
			envValue:     "",
			expected:     "default",
		},
		{
			name:         "returns default when empty string",
			key:          "EMPTY_VAR",
			defaultValue: "default",
			envValue:     "",
			expected:     "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original
			original := os.Getenv(tt.key)
			defer func() {
				if original != "" {
					os.Setenv(tt.key, original)
				} else {
					os.Unsetenv(tt.key)
				}
			}()

			// Set test value
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
			} else {
				os.Unsetenv(tt.key)
			}

			// Test
			result := getEnv(tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper functions

func saveEnv() map[string]string {
	vars := []string{
		"DB_HOST", "DB_PORT", "DB_NAME", "DB_USER", "DB_PASSWORD",
		"DATABASE_URL", "PORT", "FRONTEND_URL", "JWT_SECRET",
		"AWS_REGION", "S3_BUCKET", "ENVIRONMENT", "BANK_INFO_ENCRYPTION_KEY",
	}

	saved := make(map[string]string)
	for _, v := range vars {
		saved[v] = os.Getenv(v)
	}
	return saved
}

func restoreEnv(env map[string]string) {
	// Clear all test vars first
	for k := range env {
		os.Unsetenv(k)
	}

	// Restore original values
	for k, v := range env {
		if v != "" {
			os.Setenv(k, v)
		}
	}
}
