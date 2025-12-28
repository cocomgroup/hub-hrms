package api

import (
	"encoding/json"
	"net/http"
	"log"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"hub-hrms/backend/internal/models"
	"hub-hrms/backend/internal/service"
)

// RegisterUserRoutes registers all user management routes
func RegisterUserRoutes(r chi.Router, services *service.Services) {
	r.Route("/users", func(r chi.Router) {
		// Apply auth middleware to all user routes
		r.Use(authMiddleware(services))
		
		// User CRUD endpoints
		r.Get("/", listUsersHandler(services))
		r.Post("/", createUserHandler(services))
		r.Get("/{id}", getUserHandler(services))
		r.Put("/{id}", updateUserHandler(services))
		r.Delete("/{id}", deleteUserHandler(services))
		r.Post("/{id}/reset-password", resetPasswordHandler(services))
	})
}

// CreateUserRequest for creating new users
type CreateUserRequest struct {
	Email      string     `json:"email" validate:"required,email"`
	Username   string     `json:"username" validate:"required"`
	Password   string     `json:"password" validate:"required,min=8"`
	Role       string     `json:"role" validate:"required"`
	EmployeeID *uuid.UUID `json:"employee_id,omitempty"`
}

// UpdateUserRequest for updating users
type UpdateUserRequest struct {
	Email      *string    `json:"email,omitempty"`
	Role       *string    `json:"role,omitempty"`
	EmployeeID *uuid.UUID `json:"employee_id,omitempty"`
}

// ResetPasswordRequest for resetting passwords
type ResetPasswordRequest struct {
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

// ===========================================
// HANDLERS
// ===========================================

// createUserHandler creates a new user with encrypted password
func createUserHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only admins can create users
		// TODO: Add role check middleware
		
		var req CreateUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("ERROR: invalid request body: %v", err)
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		// Validate required fields
		if req.Username == "" {
			log.Printf("ERROR: username is required")
			respondError(w, http.StatusBadRequest, "username is required")
			return
		}
		if req.Email == "" {
			log.Printf("ERROR: email is required")
			respondError(w, http.StatusBadRequest, "email is required")
			return
		}
		if req.Password == "" {
			log.Printf("ERROR: password is required")
			respondError(w, http.StatusBadRequest, "password is required")
			return
		}
		if len(req.Password) < 8 {
			respondError(w, http.StatusBadRequest, "password must be at least 8 characters")
			return
		}
		if req.Role == "" {
			log.Printf("ERROR: role is required")
			respondError(w, http.StatusBadRequest, "role is required")
			return
		}

		// Validate role
		validRoles := map[string]bool{
			"admin":       true,
			"hr-manager":  true,
			"manager":     true,
			"employee":    true,
		}
		if !validRoles[req.Role] {
			respondError(w, http.StatusBadRequest, "invalid role (must be: admin, hr-manager, manager, or employee)")
			return
		}

		// Hash the password using auth service
		hashedPassword, err := services.Auth.HashPassword(req.Password)
		if err != nil {
			log.Printf("ERROR: failed to encrypt password: %v", err)
			respondError(w, http.StatusInternalServerError, "failed to encrypt password")
			return
		}

		// Create user model
		user := &models.User{
			Username:     req.Username,
			Email:        req.Email,
			PasswordHash: hashedPassword,
			Role:         req.Role,
			EmployeeID:   req.EmployeeID,
		}

		// Save to database
		if err := services.User.Create(r.Context(), user); err != nil {
			log.Printf("ERROR: failed to create user: %v", err)
			respondError(w, http.StatusInternalServerError, "failed to create user")
			return
		}

		// Return user (without password hash)
		user.PasswordHash = "" // Don't expose password hash
		respondJSON(w, http.StatusCreated, user)
	}
}

// listUsersHandler lists all users with optional search and role filter
func listUsersHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get query parameters
		search := r.URL.Query().Get("search")
		role := r.URL.Query().Get("role")

		// Get users from service
		users, err := services.User.List(r.Context(), search, role)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to list users")
			return
		}

		// Remove password hashes from response
		for _, user := range users {
			user.PasswordHash = ""
		}

		respondJSON(w, http.StatusOK, users)
	}
}

// getUserHandler gets a specific user by ID
func getUserHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid user ID")
			return
		}

		user, err := services.User.GetByID(r.Context(), id)
		if err != nil {
			respondError(w, http.StatusNotFound, "user not found")
			return
		}

		// Don't expose password hash
		user.PasswordHash = ""
		respondJSON(w, http.StatusOK, user)
	}
}

// updateUserHandler updates a user
func updateUserHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid user ID")
			return
		}

		// Get existing user
		user, err := services.User.GetByID(r.Context(), id)
		if err != nil {
			respondError(w, http.StatusNotFound, "user not found")
			return
		}

		// Parse update request
		var req UpdateUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		// Update fields if provided
		if req.Email != nil {
			user.Email = *req.Email
		}
		if req.Role != nil {
			// Validate role
			validRoles := map[string]bool{
				"admin":       true,
				"hr-manager":  true,
				"manager":     true,
				"employee":    true,
			}
			if !validRoles[*req.Role] {
				respondError(w, http.StatusBadRequest, "invalid role")
				return
			}
			user.Role = *req.Role
		}
		if req.EmployeeID != nil {
			user.EmployeeID = req.EmployeeID
		}

		// Save updates
		if err := services.User.Update(r.Context(), user); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to update user")
			return
		}

		// Don't expose password hash
		user.PasswordHash = ""
		respondJSON(w, http.StatusOK, user)
	}
}

// deleteUserHandler deletes a user
func deleteUserHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid user ID")
			return
		}

		// Delete user
		if err := services.User.Delete(r.Context(), id); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to delete user")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// resetPasswordHandler resets a user's password
func resetPasswordHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid user ID")
			return
		}

		// Parse request
		var req ResetPasswordRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		// Validate password
		if req.NewPassword == "" {
			respondError(w, http.StatusBadRequest, "new_password is required")
			return
		}
		if len(req.NewPassword) < 8 {
			respondError(w, http.StatusBadRequest, "password must be at least 8 characters")
			return
		}

		// Get user
		user, err := services.User.GetByID(r.Context(), id)
		if err != nil {
			respondError(w, http.StatusNotFound, "user not found")
			return
		}

		// Hash new password
		hashedPassword, err := services.Auth.HashPassword(req.NewPassword)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to encrypt password")
			return
		}

		// Update password
		user.PasswordHash = hashedPassword
		if err := services.User.Update(r.Context(), user); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to update password")
			return
		}

		respondJSON(w, http.StatusOK, map[string]string{
			"message": "Password reset successfully",
		})
	}
}