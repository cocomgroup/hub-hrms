package models

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// User represents an authenticated user
type User struct {
	ID           uuid.UUID  `json:"id"`
	Username     string     `json:"username"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"`
	Role         string     `json:"role"`
	EmployeeID   *uuid.UUID `json:"employee_id,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}



// Request/Response DTOs

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token     string    `json:"token"`
	User      User      `json:"user"`
	Employee  *Employee `json:"employee,omitempty"`
}

// StringArray is a custom type for PostgreSQL text[] arrays
// It implements sql.Scanner and driver.Valuer for robust array handling
type StringArray []string

// Scan implements sql.Scanner interface
// This handles reading array data from PostgreSQL
func (a *StringArray) Scan(src interface{}) error {
	if src == nil {
		*a = StringArray{}
		return nil
	}

	// Handle different input types
	switch v := src.(type) {
	case []byte:
		// Parse PostgreSQL array format: {item1,item2,item3}
		return a.scanBytes(v)
	case string:
		// Parse string representation
		return a.scanBytes([]byte(v))
	case []string:
		// Direct slice assignment
		*a = StringArray(v)
		return nil
	default:
		return fmt.Errorf("unsupported type for StringArray: %T", src)
	}
}

// scanBytes parses PostgreSQL array format
func (a *StringArray) scanBytes(src []byte) error {
	str := string(src)
	
	// Handle empty or NULL
	if str == "" || str == "NULL" {
		*a = StringArray{}
		return nil
	}

	// Handle empty array {}
	if str == "{}" {
		*a = StringArray{}
		return nil
	}

	// Remove outer braces
	if strings.HasPrefix(str, "{") && strings.HasSuffix(str, "}") {
		str = str[1 : len(str)-1]
	} else {
		// Invalid format - return empty array instead of error for robustness
		*a = StringArray{}
		return nil
	}

	// Handle still empty after removing braces
	if str == "" {
		*a = StringArray{}
		return nil
	}

	// Split by comma and clean up
	parts := strings.Split(str, ",")
	result := make([]string, 0, len(parts))
	
	for _, part := range parts {
		// Trim quotes and whitespace
		part = strings.TrimSpace(part)
		part = strings.Trim(part, `"`)
		if part != "" {
			result = append(result, part)
		}
	}

	*a = StringArray(result)
	return nil
}

// Value implements driver.Valuer interface
// This handles writing array data to PostgreSQL
func (a StringArray) Value() (driver.Value, error) {
	if a == nil || len(a) == 0 {
		return pq.Array([]string{}), nil
	}
	return pq.Array([]string(a)), nil
}

// ToSlice converts StringArray to []string
func (a StringArray) ToSlice() []string {
	if a == nil {
		return []string{}
	}
	return []string(a)
}

// FromSlice creates StringArray from []string
func FromSlice(s []string) StringArray {
	if s == nil {
		return StringArray{}
	}
	return StringArray(s)
}
