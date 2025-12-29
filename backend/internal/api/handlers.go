package api

import (
	"context"
	"encoding/json"
	"fmt"
	"hub-hrms/backend/internal/models"
	"hub-hrms/backend/internal/service"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// RegisterAuthRoutes registers authentication routes
func RegisterAuthRoutes(r chi.Router, services *service.Services) {
	r.Post("/auth/login", loginHandler(services))
}


// Middleware - FIXED VERSION
func authMiddleware(services *service.Services) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				respondError(w, http.StatusUnauthorized, "missing authorization header")
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				respondError(w, http.StatusUnauthorized, "invalid authorization header")
				return
			}

			token, err := services.Auth.ValidateToken(parts[1])
			if err != nil || !token.Valid {
				respondError(w, http.StatusUnauthorized, "invalid token")
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				respondError(w, http.StatusUnauthorized, "invalid token claims")
				return
			}

			// Debug: log claims
			log.Printf("DEBUG JWT Middleware: Claims = %+v", claims)

			// Add both user_id and claims to context
			ctx := r.Context()
			
			// Set user_id in context (as string)
			if userID, ok := claims["user_id"].(string); ok {
				log.Printf("DEBUG JWT Middleware: Setting user_id in context: %s", userID)
				ctx = context.WithValue(ctx, "user_id", userID)
			} else {
				log.Printf("WARNING JWT Middleware: user_id not found in claims or wrong type")
			}
			
			// ADDED: Set claims in context so other functions can access them
			ctx = context.WithValue(ctx, "claims", claims)
			log.Printf("DEBUG JWT Middleware: Set claims in context")

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Auth handlers
func loginHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		response, err := services.Auth.Login(r.Context(), &req)
		if err != nil {
			respondError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}

		respondJSON(w, http.StatusOK, response)
	}
}

// Helper functions - FIXED VERSION

// getUserIDFromContext extracts user ID from context (returns UUID)
func getUserIDFromContext(ctx context.Context) (uuid.UUID, error) {
	userIDStr, ok := ctx.Value("user_id").(string)
	if !ok {
		log.Printf("DEBUG getUserIDFromContext: user_id not found in context")
		return uuid.Nil, fmt.Errorf("user_id not found in context")
	}
	log.Printf("DEBUG getUserIDFromContext: Found user_id: %s", userIDStr)
	return uuid.Parse(userIDStr)
}

// getEmployeeIDFromContext extracts employee/user ID from context (returns UUID)
// FIXED: Now tries user_id first, then falls back to claims
func getEmployeeIDFromContext(ctx context.Context) (uuid.UUID, error) {
	// PRIORITY: Try to get employee_id from claims FIRST
	// This is the correct field for time entries, PTO, etc.
	
	// Try jwt.MapClaims type first
	if claims, ok := ctx.Value("claims").(jwt.MapClaims); ok {
		// Try employee_id from claims (CORRECT for most employee operations)
		if employeeIDStr, ok := claims["employee_id"].(string); ok {
			log.Printf("DEBUG getEmployeeIDFromContext: Found employee_id in claims: %s", employeeIDStr)
			return uuid.Parse(employeeIDStr)
		}
		log.Printf("DEBUG getEmployeeIDFromContext: employee_id not found in jwt.MapClaims")
	} else {
		log.Printf("DEBUG getEmployeeIDFromContext: jwt.MapClaims not found, trying map[string]interface{}")
		
		// Try map[string]interface{} as fallback
		if claims, ok := ctx.Value("claims").(map[string]interface{}); ok {
			if employeeIDStr, ok := claims["employee_id"].(string); ok {
				log.Printf("DEBUG getEmployeeIDFromContext: Found employee_id in map claims: %s", employeeIDStr)
				return uuid.Parse(employeeIDStr)
			}
			log.Printf("DEBUG getEmployeeIDFromContext: employee_id not found in map claims")
		} else {
			log.Printf("DEBUG getEmployeeIDFromContext: Claims not found in context")
		}
	}
	
	// Fallback: try user_id from context (for backwards compatibility)
	// This is only correct if user_id == employee_id (which is wrong design)
	if userIDStr, ok := ctx.Value("user_id").(string); ok {
		log.Printf("DEBUG getEmployeeIDFromContext: Falling back to user_id in context: %s", userIDStr)
		employeeID, err := uuid.Parse(userIDStr)
		if err == nil {
			return employeeID, nil
		}
		log.Printf("DEBUG getEmployeeIDFromContext: Failed to parse user_id: %v", err)
	}

	log.Printf("DEBUG getEmployeeIDFromContext: No valid employee ID found")
	return uuid.Nil, fmt.Errorf("unauthorized: employee ID not found")
}
func getUserIDFromJWTContext(r *http.Request) uuid.UUID {
	// Debug: log what's in context
	log.Printf("DEBUG: Attempting to get user_id from context")
	
	// This assumes the JWT middleware sets the user_id in context
	if userID := r.Context().Value("user_id"); userID != nil {
		log.Printf("DEBUG: Found user_id in context: %v (type: %T)", userID, userID)
		if id, ok := userID.(uuid.UUID); ok {
			log.Printf("DEBUG: Successfully parsed as uuid.UUID: %s", id)
			return id
		}
		if idStr, ok := userID.(string); ok {
			log.Printf("DEBUG: user_id is string: %s", idStr)
			if id, err := uuid.Parse(idStr); err == nil {
				log.Printf("DEBUG: Successfully parsed string to UUID: %s", id)
				return id
			} else {
				log.Printf("DEBUG: Failed to parse string to UUID: %v", err)
			}
		}
	} else {
		log.Printf("DEBUG: user_id not found in context")
	}
	
	// Fallback: try to get from claims in context
	if claims := r.Context().Value("claims"); claims != nil {
		log.Printf("DEBUG: Found claims in context")
		if claimsMap, ok := claims.(map[string]interface{}); ok {
			if userIDStr, ok := claimsMap["user_id"].(string); ok {
				if id, err := uuid.Parse(userIDStr); err == nil {
					log.Printf("DEBUG: Got user_id from claims: %s", id)
					return id
				}
			}
		}
	}
	
	// Return nil UUID as fallback (should be handled by auth middleware)
	log.Printf("WARNING: Returning nil UUID - user_id not found in context")
	return uuid.Nil
}

// Helper functions for responses
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}