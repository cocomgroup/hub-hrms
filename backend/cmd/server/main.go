package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
    "path/filepath"
    "strings"

	"hub-hrms/backend/internal/api"
    "hub-hrms/backend/internal/db"
	"hub-hrms/backend/internal/config"
	"hub-hrms/backend/internal/repository"
	"hub-hrms/backend/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func main() {

    cfg := config.Load()

	log.Printf("Starting backend server...")

	log.Printf("Database URL: %s", maskPassword(cfg.DatabaseURL))

	// Connect to database with retries
	var database *db.Postgres
	var err error

	for i := 0; i < 30; i++ {
		database, err = db.NewPostgres(cfg.DatabaseURL)
		if err == nil {
			log.Printf("Connected to database")
			break
		}
		log.Printf("Database connection attempt %d/30: %v", i+1, err)
		time.Sleep(5 * time.Second)
	}
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Run migrations
	log.Printf("Running database migrations...")
   
	if err := runMigrations(database); err != nil {
		log.Printf("WARNING: Migration error: %v", err)
	} else {
		log.Printf("Migrations completed")
	}


	// Initialize repositories
    var dbPool = database.GetPool() 
	repos := repository.NewRepositories(dbPool)
    runMigrations(database)

	// Initialize services
	services := service.NewServices(repos, cfg)

	// Initialize HTTP server
	router := setupRouter(services, cfg)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Starting server on port %s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

func setupRouter(services *service.Services, cfg *config.Config) *chi.Mux {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))


	// API routes
	r.Route("/api", func(r chi.Router) {
        
		// Health check
        r.Route("/health", func(r chi.Router) {
            r.Get("/", healthCheckHandler)
            r.Head("/", healthCheckHandler)
        })
		
		api.RegisterAuthRoutes(r, services)
        api.RegisterUserRoutes(r, services)
		api.RegisterEmployeeRoutes(r, services)
		api.RegisterOnboardingRoutes(r, services)
		api.RegisterWorkflowRoutes(r, services)
		api.RegisterTimesheetRoutes(r, services)
		api.RegisterPTORoutes(r, services)
		api.RegisterBenefitsRoutes(r, services)
		api.RegisterPayrollRoutes(r, services)
        api.RegisterRecruitingRoutes(r, services)
        api.RegisterOrganizationRoutes(r, services)
	})

	return r
}

// Health function:
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    
    // Only send body for GET (HEAD must have no body per HTTP spec)
    if r.Method == http.MethodGet {
        w.Write([]byte(`{"status":"healthy","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`))
    }
}
type Config struct {
	DatabaseURL    string
	ServerAddr     string
    Port           string
    JWTSecret      string
    FrontendURL    string
    Environment    string
	S3Bucket       string
}

type DatabaseSecret struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     string `json:"port"`
	DBName   string `json:"dbname"`
}


func maskPassword(connStr string) string {
	result := connStr
	if idx := len("postgres://"); idx < len(connStr) {
		afterUser := connStr[idx:]
		if colonIdx := indexOf(afterUser, ":"); colonIdx > 0 {
			if atIdx := indexOf(afterUser, "@"); atIdx > colonIdx {
				result = connStr[:idx+colonIdx+1] + "****" + connStr[idx+atIdx:]
			}
		}
	}
	return result
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func runMigrations(database *db.Postgres) error {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Migration files directory
	migrationsDir := "cmd/migrations"

	// Get all SQL files in order
	migrationFiles := []string{
		"001-create-users.sql",
		"002-create-employees.sql",
		"003-create-onboarding-tasks.sql",
		"004-create-time-sheet-tables.sql",
		"005-create-benefits-table.sql",
		"006-create-pto-tables.sql",
		"007-create-workflows.sql",
		"008-create-payroll.sql",
		"009-create-recruiting.sql",
		"010-create-organizations.sql",
        "011-seed-users.sql",
        "012-seed-employees.sql",
        "013-seed-organizations.sql",
        "014-seed-recruiting.sql",
        "015-seed-pto-benefits.sql",
	}

	log.Printf("Running %d database migrations from %s/", len(migrationFiles), migrationsDir)

	for i, filename := range migrationFiles {
		migrationPath := filepath.Join(migrationsDir, filename)
		
		log.Printf("  [%d/%d] Running %s...", i+1, len(migrationFiles), filename)

		// Read SQL file
		sqlBytes, err := os.ReadFile(migrationPath)
		if err != nil {
			// If file doesn't exist, log warning and continue
			if os.IsNotExist(err) {
				log.Printf("  ⚠️  WARNING: Migration file not found: %s (skipping)", migrationPath)
				continue
			}
			return fmt.Errorf("failed to read migration %s: %w", filename, err)
		}

		sql := string(sqlBytes)

		// Skip empty files
		if len(strings.TrimSpace(sql)) == 0 {
			log.Printf("  ⏭️  Skipping empty migration: %s", filename)
			continue
		}

		// Execute migration
		if err := database.Exec(ctx, sql); err != nil {
			return fmt.Errorf("migration %s failed: %w", filename, err)
		}

		log.Printf("  ✅ Completed %s", filename)
	}

	log.Printf("✅ All migrations completed successfully!")
	return nil
}
