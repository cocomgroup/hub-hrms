package middleware

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"strings"
	"time"
)

type contextKey string

const (
	APIKeyIDKey contextKey = "api_key_id"
	ScopesKey   contextKey = "scopes"
	APIKeyName  contextKey = "api_key_name"
)

// APIKeyAuth validates API keys from the database
func APIKeyAuth(db *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startTime := time.Now()

			// Extract API key from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				respondError(w, "Missing Authorization header", http.StatusUnauthorized)
				return
			}

			// Parse "Bearer <token>"
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				respondError(w, "Invalid Authorization format. Use: Bearer <token>", http.StatusUnauthorized)
				return
			}

			apiKey := parts[1]

			// Verify API key using PostgreSQL function
			var (
				apiKeyID string
				name     string
				scopes   []string
				isValid  bool
			)

			query := `SELECT api_key_id, name, scopes, is_valid FROM verify_api_key($1)`
			err := db.QueryRowContext(r.Context(), query, apiKey).Scan(&apiKeyID, &name, &scopes, &isValid)

			if err == sql.ErrNoRows || !isValid {
				log.Printf("Invalid API key attempt from %s", r.RemoteAddr)
				respondError(w, "Invalid or expired API key", http.StatusUnauthorized)
				
				// Log failed attempt (optional)
				logAPIKeyUsage(db, r.Context(), "", r.URL.Path, r.Method, 401, 
					r.RemoteAddr, r.UserAgent(), int(time.Since(startTime).Milliseconds()))
				return
			}
			if err != nil {
				log.Printf("Error verifying API key: %v", err)
				respondError(w, "Authentication error", http.StatusInternalServerError)
				return
			}

			// Add to context
			ctx := context.WithValue(r.Context(), APIKeyIDKey, apiKeyID)
			ctx = context.WithValue(r.Context(), ScopesKey, scopes)
			ctx = context.WithValue(r.Context(), APIKeyName, name)

			log.Printf("API key authenticated: %s (name: %s)", apiKeyID, name)

			// Wrap response writer to capture status code
			wrapped := &responseWriter{ResponseWriter: w, statusCode: 200}

			// Continue to next handler
			next.ServeHTTP(wrapped, r.WithContext(ctx))

			// Log API usage
			duration := int(time.Since(startTime).Milliseconds())
			go logAPIKeyUsage(db, context.Background(), apiKeyID, r.URL.Path, r.Method, 
				wrapped.statusCode, r.RemoteAddr, r.UserAgent(), duration)
		})
	}
}

// JWTAuth validates JWT tokens for internal API (REST endpoints)
func JWTAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract JWT from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			respondError(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}

		// Parse "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			respondError(w, "Invalid Authorization format", http.StatusUnauthorized)
			return
		}

		token := parts[1]

		// TODO: Validate JWT token
		// For now, just pass through
		// In production, verify JWT signature and claims
		_ = token

		next.ServeHTTP(w, r)
	})
}

// Helper functions

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func respondError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write([]byte(`{"errors":[{"message":"` + message + `"}]}`))
}

func logAPIKeyUsage(db *sql.DB, ctx context.Context, apiKeyID, endpoint, method string, 
	statusCode int, remoteAddr, userAgent string, responseTime int) {
	
	if apiKeyID == "" {
		return // Don't log failed auth attempts with no key
	}

	// Extract IP from remoteAddr (remove port)
	ip := strings.Split(remoteAddr, ":")[0]

	_, err := db.ExecContext(ctx, `
		SELECT log_api_key_usage($1, $2, $3, $4, $5::inet, $6, $7)
	`, apiKeyID, endpoint, method, statusCode, ip, userAgent, responseTime)

	if err != nil {
		log.Printf("Failed to log API usage: %v", err)
	}
}

// HasScope checks if the API key has a specific scope
func HasScope(ctx context.Context, requiredScope string) bool {
	scopes, ok := ctx.Value(ScopesKey).([]string)
	if !ok {
		return false
	}

	for _, scope := range scopes {
		if scope == requiredScope {
			return true
		}
	}
	return false
}

// GetAPIKeyID retrieves the API key ID from context
func GetAPIKeyID(ctx context.Context) (string, bool) {
	apiKeyID, ok := ctx.Value(APIKeyIDKey).(string)
	return apiKeyID, ok
}