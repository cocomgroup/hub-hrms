package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"hub-hrms/backend/internal/models"
)

// UserRepository interface
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	List(ctx context.Context, search string, role string) ([]*models.User, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// UserRepository implementation
type userRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (username, email, password_hash, role, employee_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(ctx, query,
		user.Username, user.Email, user.PasswordHash, user.Role, user.EmployeeID,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, username, email, password_hash, role, employee_id, created_at, updated_at
		FROM users WHERE email = $1
	`
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role,
		&user.EmployeeID, &user.CreatedAt, &user.UpdatedAt,
	)
	return user, err
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, username, email, password_hash, role, employee_id, created_at, updated_at
		FROM users WHERE id = $1
	`
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role,
		&user.EmployeeID, &user.CreatedAt, &user.UpdatedAt,
	)
	return user, err
}

func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users
		SET username = $1, email = $2, password_hash = $3, role = $4, employee_id = $5, updated_at = NOW()
		WHERE id = $6
		RETURNING updated_at
	`
	return r.db.QueryRow(ctx, query,
		user.Username, user.Email, user.PasswordHash, user.Role, user.EmployeeID, user.ID,
	).Scan(&user.UpdatedAt)
}

// List returns all users with optional search and role filter
func (r *userRepository) List(ctx context.Context, search string, role string) ([]*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, role, employee_id, created_at, updated_at
		FROM users
		WHERE 1=1
	`
	args := []interface{}{}
	argPos := 1

	// Add search filter
	if search != "" {
		query += fmt.Sprintf(" AND (email ILIKE $%d OR username ILIKE $%d)", argPos, argPos+1)
		searchPattern := "%" + search + "%"
		args = append(args, searchPattern, searchPattern)
		argPos += 2
	}

	// Add role filter
	if role != "" {
		query += fmt.Sprintf(" AND role = $%d", argPos)
		args = append(args, role)
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []*models.User{}
	for rows.Next() {
		user := &models.User{}
		err := rows.Scan(
			&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role,
			&user.EmployeeID, &user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, rows.Err()
}

func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}