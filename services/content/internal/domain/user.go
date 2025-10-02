package domain

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	Role         UserRole  `json:"role"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// UserRole represents user roles
type UserRole string

const (
	UserRoleAdmin  UserRole = "admin"
	UserRoleEditor UserRole = "editor"
	UserRoleAuthor UserRole = "author"
	UserRoleViewer UserRole = "viewer"
)

// RegisterUserInput represents input for user registration
type RegisterUserInput struct {
	Email    string   `json:"email" validate:"required,email"`
	Username string   `json:"username" validate:"required,min=3,max=50"`
	Password string   `json:"password" validate:"required,min=8,max=100"`
	Role     UserRole `json:"role" validate:"omitempty,oneof=admin editor author viewer"`
}

// LoginInput represents input for user login
type LoginInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// AuthResponse represents authentication response
type AuthResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	User      *User     `json:"user"`
}

// HasPermission checks if user has permission for an action
func (u *User) HasPermission(action string) bool {
	switch u.Role {
	case UserRoleAdmin:
		return true
	case UserRoleEditor:
		return action == "read" || action == "create" || action == "update"
	case UserRoleAuthor:
		return action == "read" || action == "create"
	case UserRoleViewer:
		return action == "read"
	default:
		return false
	}
}

// CanEditContent checks if user can edit specific content
func (u *User) CanEditContent(content *Content) bool {
	if u.Role == UserRoleAdmin || u.Role == UserRoleEditor {
		return true
	}
	if u.Role == UserRoleAuthor && content.AuthorID == u.ID {
		return true
	}
	return false
}
